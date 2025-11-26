package executor

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/platforms"
)

// ActionExecution represents an action being executed
type ActionExecution struct {
	Index      int
	Action     config.Action
	App        config.AppConfig
	Platform   platforms.Platform
	Recording  *string
	Result     *TestResult
	Error      error
}

// ActionGroup represents a group of actions that can be executed in parallel
type ActionGroup struct {
	Actions []config.Action
	// True if actions in this group can be parallelized
	Parallelizable bool
}

// analyzeActions determines which actions can be parallelized
func (e *Executor) analyzeActions() []ActionGroup {
	if len(e.config.Actions) == 0 {
		return []ActionGroup{}
	}
	
	groups := make([]ActionGroup, 0)
	currentGroup := ActionGroup{
		Actions:      make([]config.Action, 0),
		Parallelizable: true,
	}
	
	for _, action := range e.config.Actions {
		// Actions that modify state or depend on order must be sequential
		needsSequential := action.Type == "navigate" ||
			action.Type == "submit" ||
			action.Type == "record" && action.Duration > 0 ||
			action.Type == "wait" && action.WaitTime > 2 // long waits separate groups
		
		if needsSequential {
			// Add current group if not empty
			if len(currentGroup.Actions) > 0 {
				groups = append(groups, currentGroup)
			}
			
			// Start new group with this action
			groups = append(groups, ActionGroup{
				Actions:      []config.Action{action},
				Parallelizable: false,
			})
			
			// Reset current group
			currentGroup = ActionGroup{
				Actions:      make([]config.Action, 0),
				Parallelizable: true,
			}
		} else {
			// Add to current parallelizable group
			currentGroup.Actions = append(currentGroup.Actions, action)
		}
	}
	
	// Add final group if not empty
	if len(currentGroup.Actions) > 0 {
		groups = append(groups, currentGroup)
	}
	
	return groups
}

// executeActionGroup executes a group of actions
func (e *Executor) executeActionGroup(ctx context.Context, platform platforms.Platform, app config.AppConfig, group ActionGroup, result *TestResult, recordingFile *string) error {
	if !group.Parallelizable || len(group.Actions) <= 1 {
		// Execute sequentially
		for _, action := range group.Actions {
			if err := e.executeAction(platform, action, app, result, recordingFile); err != nil {
				return fmt.Errorf("action '%s' failed: %w", action.Name, err)
			}
		}
		return nil
	}
	
	// Execute in parallel with controlled goroutine count
	maxWorkers := runtime.NumCPU()
	if maxWorkers > 4 {
		maxWorkers = 4 // Limit parallelism for stability
	}
	
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(group.Actions))
	
	for _, action := range group.Actions {
		wg.Add(1)
		
		go func(act config.Action) {
			defer wg.Done()
			
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
				
				// Clone platform for goroutine safety if needed
				if err := e.executeAction(platform, act, app, result, recordingFile); err != nil {
					errorChan <- fmt.Errorf("action '%s' failed: %w", act.Name, err)
				}
			case <-ctx.Done():
				errorChan <- ctx.Err()
			}
		}(action)
	}
	
	// Wait for all actions to complete
	go func() {
		wg.Wait()
		close(errorChan)
	}()
	
	// Check for errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}
	
	return nil
}

// executeAppParallel executes app actions with parallelization optimization
func (e *Executor) executeAppParallel(ctx context.Context, app config.AppConfig) TestResult {
	startTime := time.Now()
	
	result := TestResult{
		AppName:    app.Name,
		AppType:    app.Type,
		StartTime:  startTime,
		Metrics:    make(map[string]interface{}),
		Screenshots: make([]string, 0),
		Videos:      make([]string, 0),
		Success:    false,
	}
	
	// Create platform
	platform, err := e.factory.CreatePlatform(app.Type)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create platform: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}
	
	// Initialize platform
	if err := platform.Initialize(app); err != nil {
		result.Error = fmt.Sprintf("Failed to initialize platform: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}
	
	defer platform.Close()
	
	// Analyze actions for parallelization opportunities
	groups := e.analyzeActions()
	e.logger.Debugf("Analyzing %d actions into %d groups for parallel execution", len(e.config.Actions), len(groups))
	
	// Execute action groups sequentially, but within each group execute in parallel when possible
	currentRecordingFile := ""
	for i, group := range groups {
		e.logger.Debugf("Executing action group %d: %d actions, parallelizable: %v", i, len(group.Actions), group.Parallelizable)
		
		if err := e.executeActionGroup(ctx, platform, app, group, &result, &currentRecordingFile); err != nil {
			result.Error = err.Error()
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		}
		
		// Check context cancellation between groups
		select {
		case <-ctx.Done():
			result.Error = "Execution cancelled"
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		default:
		}
	}
	
	// Stop recording if still active
	if currentRecordingFile != "" {
		if err := platform.StopRecording(); err != nil {
			e.logger.Errorf("Failed to stop recording: %v", err)
		}
	}
	
	// Get final metrics
	result.Metrics = platform.GetMetrics()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = true
	
	e.logger.Infof("executeAppParallel completed successfully for %s", app.Name)
	
	return result
}