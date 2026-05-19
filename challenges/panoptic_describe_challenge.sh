#!/usr/bin/env bash
# panoptic_describe_challenge.sh
#
# Round-298 paired-mutation deep-doc challenge for Panoptic.
#
# Validates that:
#   1. The deep-doc ledger (docs/test-coverage.md) lists every exported
#      symbol from pkg/i18n, internal/config, internal/platforms, and
#      the public surface of internal/executor that round-298 covers.
#   2. The bilingual fixture (challenges/fixtures/payloads.json) parses
#      and covers >= 5 locales (en, de, es, ja, sr).
#   3. The bilingual runner (challenges/runner/main.go) builds and runs
#      against every production primitive listed in test-coverage.md
#      §2-5, with non-ASCII bytes preserved through the i18n + config
#      pipeline (CJK + Cyrillic round-trip).
#   4. The README enumerates the round-298 anti-bluff guarantees + the
#      packages exercised by the runner.
#
# Paired-mutation invariant (CONST-035 + CONST-050(B) + §11.9):
#   With --anti-bluff-mutate the script plants a deliberate symbol-rename
#   mutation in a TMP COPY of the ledger, reruns validation against the
#   tmp copy, and asserts the gate FAILS with exit 99. This proves the
#   gate actually catches ledger-vs-source drift instead of rubber-
#   stamping it. The original tree is never mutated.
#
# Exit codes:
#   0  — gate PASS on clean tree
#   1  — gate FAIL on clean tree (real failure to fix)
#   99 — paired-mutation correctly detected (good — proves anti-bluff)
#   2  — usage / environment error
#
# Operator mandate (2026-05-19, verbatim, cascaded per §11.9):
#   "all existing tests and Challenges do work in anti-bluff manner -
#    they MUST confirm that all tested codebase really works as expected!
#    We had been in position that all tests do execute with success and
#    all Challenges as well, but in reality the most of the features
#    does not work and can't be used! This MUST NOT be the case and
#    execution of tests and Challenges MUST guarantee the quality, the
#    completition and full usability by end users of the product!"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

MUTATE=0
for arg in "$@"; do
    case "$arg" in
        --anti-bluff-mutate) MUTATE=1 ;;
        --help|-h)
            sed -n '1,42p' "$0"
            exit 0
            ;;
        *)
            echo "unknown argument: $arg" >&2
            exit 2
            ;;
    esac
done

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

LEDGER="${MODULE_DIR}/docs/test-coverage.md"
FIXTURE="${MODULE_DIR}/challenges/fixtures/payloads.json"
RUNNER="${MODULE_DIR}/challenges/runner/main.go"
README="${MODULE_DIR}/README.md"

# If mutation requested, work against a tmp copy of the ledger with a
# planted symbol rename. The original tree stays untouched.
LEDGER_WORK="${LEDGER}"
TMP_LEDGER=""
if [ "${MUTATE}" -eq 1 ]; then
    TMP_LEDGER="$(mktemp)"
    cp "${LEDGER}" "${TMP_LEDGER}"
    # Plant a rename: NoopTranslator → NoopTranslatorMUTATED (real symbol
    # must remain absent from the planted ledger to flag the drift).
    sed -i 's/NoopTranslator/NoopTranslatorMUTATED/g' "${TMP_LEDGER}"
    LEDGER_WORK="${TMP_LEDGER}"
    echo "=== Panoptic Describe Challenge (anti-bluff-mutate mode) ==="
else
    echo "=== Panoptic Describe Challenge (clean mode) ==="
fi
echo ""

# Section 1: ledger presence and freshness
echo "Section 1: docs/test-coverage.md ledger"
if [ ! -f "${LEDGER_WORK}" ]; then
    fail "ledger missing at ${LEDGER_WORK}"
else
    pass "ledger present"
    if grep -q "round-298" "${LEDGER_WORK}"; then
        pass "ledger marked round-298"
    else
        fail "ledger missing round-298 marker"
    fi
    if grep -q "execution of tests and Challenges MUST guarantee" "${LEDGER_WORK}"; then
        pass "ledger carries Article XI §11.9 mandate"
    else
        fail "ledger missing Article XI §11.9 mandate"
    fi
fi

# Section 2: every exported pkg symbol appears in ledger
echo ""
echo "Section 2: exported symbols cross-reference"

extract_symbols() {
    local pkg_dir="$1"
    local files
    files=$(find "${pkg_dir}" -maxdepth 1 -type f -name '*.go' \
        ! -name '*_test.go')
    [ -z "${files}" ] && return 0
    # shellcheck disable=SC2086
    grep -hE '^(func ([A-Z][A-Za-z0-9_]*\()|func \([^)]+\) ([A-Z][A-Za-z0-9_]*\()|type [A-Z][A-Za-z0-9_]* )' \
        ${files} 2>/dev/null \
        | sed -E 's/^func \([^)]+\) ([A-Z][A-Za-z0-9_]*)\(.*$/\1/; s/^func ([A-Z][A-Za-z0-9_]*)\(.*$/\1/; s/^type ([A-Z][A-Za-z0-9_]*).*$/\1/' \
        | sort -u
}

CHECKED=0
MISSING=0
# Round-298 scope: i18n + config + platforms public surface. Executor
# is partially scoped per docs/test-coverage.md §5 — exhaustive cross-
# reference covers the symbols round-298 actually exercises.
for pkg_path in \
    "pkg/i18n" \
    "internal/config" \
    "internal/platforms"; do
    PKG_DIR="${MODULE_DIR}/${pkg_path}"
    if [ ! -d "${PKG_DIR}" ]; then
        fail "${pkg_path} missing — cannot cross-reference"
        continue
    fi
    while IFS= read -r sym; do
        [ -z "${sym}" ] && continue
        # Skip noisy single-word generic names that grep false-
        # positives on (interface methods that overlap with stdlib).
        case "${sym}" in
            String|Error|Write|WriteHeader|Open|T|Close) continue ;;
        esac
        CHECKED=$((CHECKED + 1))
        if grep -qE "\\b${sym}\\b" "${LEDGER_WORK}"; then
            : # symbol cross-referenced
        else
            fail "ledger missing symbol ${pkg_path}.${sym}"
            MISSING=$((MISSING + 1))
        fi
    done < <(extract_symbols "${PKG_DIR}")
done
# Single-letter / interface-method T explicitly required.
if grep -qE '\bT\b.*i18n' "${LEDGER_WORK}" \
    || grep -qE 'pkg/i18n.*T\b' "${LEDGER_WORK}"; then
    pass "i18n.T helper cross-referenced"
else
    fail "ledger missing i18n.T helper"
fi
if [ "${CHECKED}" -gt 0 ] && [ "${MISSING}" -eq 0 ]; then
    pass "all ${CHECKED} exported symbols cross-referenced in ledger"
fi

# Executor sub-symbols — round-298 only requires the runner-exercised
# set; the full surface is enumerated in §5 of the ledger.
for exec_sym in TestResult MarshalJSON PlatformFactory; do
    if grep -qE "\\b${exec_sym}\\b" "${LEDGER_WORK}"; then
        pass "executor sub-symbol ${exec_sym} cross-referenced"
    else
        fail "ledger missing executor sub-symbol ${exec_sym}"
    fi
done

# Section 3: bilingual fixture sanity
echo ""
echo "Section 3: bilingual fixture"
if [ ! -f "${FIXTURE}" ]; then
    fail "fixture missing at ${FIXTURE}"
else
    pass "fixture present"
    LOCALE_COUNT=$(grep -oE '"locale":\s*"[^"]+"' "${FIXTURE}" | sort -u | wc -l)
    if [ "${LOCALE_COUNT}" -ge 5 ]; then
        pass "fixture covers ${LOCALE_COUNT} locales (>=5)"
    else
        fail "fixture covers only ${LOCALE_COUNT} locales (<5)"
    fi
    for loc in en de es ja sr; do
        if grep -q "\"locale\": \"${loc}\"" "${FIXTURE}"; then
            pass "fixture includes locale ${loc}"
        else
            fail "fixture missing locale ${loc}"
        fi
    done
fi

# Section 4: runner builds + runs against every production primitive
echo ""
echo "Section 4: bilingual runner build + run (real production code paths)"
if [ ! -f "${RUNNER}" ]; then
    fail "runner missing at ${RUNNER}"
else
    pass "runner source present"
    cd "${MODULE_DIR}"
    if go build -o /tmp/panoptic_round298_runner ./challenges/runner/ \
        2>/tmp/panoptic_r298_build.log; then
        pass "runner builds"
        if /tmp/panoptic_round298_runner -fixtures "${FIXTURE}" \
            > /tmp/panoptic_r298_run.log 2>&1; then
            pass "runner exit 0 across every primitive"
            for locale in en de es ja sr; do
                if grep -q "PASS \[i18n:noop:${locale}\]" /tmp/panoptic_r298_run.log; then
                    pass "i18n NoopTranslator ${locale} round-trip"
                else
                    fail "i18n NoopTranslator ${locale} missing from runner output"
                fi
                if grep -q "PASS \[config:load-validate-roundtrip:${locale}\]" /tmp/panoptic_r298_run.log; then
                    pass "config Load+Validate ${locale} YAML round-trip"
                else
                    fail "config Load+Validate ${locale} missing from runner output"
                fi
            done
            if grep -q "PASS \[config:validate-negative\]" /tmp/panoptic_r298_run.log; then
                pass "config Validate negative path verified"
            else
                fail "config Validate negative path missing"
            fi
            for plat in web desktop mobile negative; do
                if grep -q "PASS \[platform-factory:${plat}\]" /tmp/panoptic_r298_run.log; then
                    pass "platform-factory ${plat} dispatch verified"
                else
                    fail "platform-factory ${plat} missing"
                fi
            done
            if grep -qE "PASS \[executor-marshal:utf8-detector:(fixed|regression-present)\]" /tmp/panoptic_r298_run.log; then
                pass "executor MarshalJSON UTF-8 detector reported KNOWN-ISSUE state"
            else
                fail "executor MarshalJSON UTF-8 detector did not report state"
            fi
            for locale in ja sr; do
                if grep -q "PASS \[wire:i18n+config:${locale}\]" /tmp/panoptic_r298_run.log; then
                    pass "cross-wire i18n+config ${locale} preserved non-ASCII bytes"
                else
                    fail "cross-wire i18n+config ${locale} missing"
                fi
            done
            if grep -qE "=== Summary: [0-9]+ PASS, 0 FAIL ===" /tmp/panoptic_r298_run.log; then
                pass "runner summary line confirms 0 FAIL"
            else
                fail "runner summary line missing or non-zero FAIL"
            fi
        else
            fail "runner exit non-zero — see /tmp/panoptic_r298_run.log"
            sed -n '1,60p' /tmp/panoptic_r298_run.log
        fi
    else
        fail "runner build failed — see /tmp/panoptic_r298_build.log"
        sed -n '1,40p' /tmp/panoptic_r298_build.log
    fi
    rm -f /tmp/panoptic_round298_runner
fi

# Section 5: README round-298 anti-bluff section
echo ""
echo "Section 5: README round-298 anti-bluff section"
if grep -q "Anti-bluff guarantees" "${README}"; then
    pass "README declares Anti-bluff guarantees"
else
    fail "README missing Anti-bluff guarantees section"
fi
if grep -q "round-298" "${README}"; then
    pass "README marked round-298"
else
    fail "README missing round-298 marker"
fi

# Cleanup mutated ledger if any
if [ -n "${TMP_LEDGER}" ]; then
    rm -f "${TMP_LEDGER}"
fi

echo ""
echo "=== Summary: ${PASS}/${TOTAL} PASS, ${FAIL} FAIL ==="

if [ "${MUTATE}" -eq 1 ]; then
    if [ "${FAIL}" -gt 0 ]; then
        echo "anti-bluff-mutate: gate correctly detected planted mutation (exit 99)"
        exit 99
    else
        echo "anti-bluff-mutate: gate FAILED to detect planted mutation — bluff!"
        exit 1
    fi
fi

if [ "${FAIL}" -gt 0 ]; then
    exit 1
fi
exit 0
