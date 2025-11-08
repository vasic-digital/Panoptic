#!/bin/bash

# Panoptic Test Data Generator
# Creates realistic test data, fixtures, and mock scenarios

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[DATA]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[DATA]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[DATA]${NC} $1"
}

print_error() {
    echo -e "${RED}[DATA]${NC} $1"
}

# Configuration
DATA_DIR="test_data"
FIXTURES_DIR="fixtures"
MOCK_APPS_DIR="mock_apps"
TEMPLATES_DIR="templates"

# Parse command line arguments
GENERATE_TYPE="all"
CLEAN_BEFORE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --type)
            GENERATE_TYPE="$2"
            shift 2
            ;;
        --clean)
            CLEAN_BEFORE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Test data generation options:"
            echo "  --type TYPE      Type of data to generate (all|configs|fixtures|mocks|templates)"
            echo "  --clean           Clean existing data before generation"
            echo "  --verbose         Enable verbose output"
            echo "  -h, --help       Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Initialize directories
init_directories() {
    print_status "Initializing test data directories..."
    
    if [ "$CLEAN_BEFORE" = true ]; then
        print_status "Cleaning existing test data..."
        rm -rf "$DATA_DIR" "$FIXTURES_DIR" "$MOCK_APPS_DIR" "$TEMPLATES_DIR"
    fi
    
    mkdir -p "$DATA_DIR"/{configs,fixtures,scenarios}
    mkdir -p "$FIXTURES_DIR"/{web,mobile,desktop,images}
    mkdir -p "$MOCK_APPS_DIR"/{web,mobile,desktop}
    mkdir -p "$TEMPLATES_DIR"/{configs,scenarios}
    
    print_success "Directories initialized"
}

# Generate test configurations
generate_configs() {
    print_status "Generating test configurations..."
    
    # Web application configurations
    cat > "$DATA_DIR/configs/web_basic.yaml" << 'EOF'
name: "Web Basic Test"
output: "./output/web_basic"
apps:
  - name: "Basic Web App"
    type: "web"
    url: "https://httpbin.org/html"
    timeout: 30
actions:
  - name: "navigate_basic"
    type: "navigate"
    value: "https://httpbin.org/html"
  - name: "wait_load"
    type: "wait"
    wait_time: 2
  - name: "screenshot_basic"
    type: "screenshot"
    parameters:
      filename: "web_basic.png"
settings:
  screenshot_format: "png"
  video_format: "mp4"
  quality: 80
  enable_metrics: true
EOF

    cat > "$DATA_DIR/configs/web_form.yaml" << 'EOF'
name: "Web Form Test"
output: "./output/web_form"
apps:
  - name: "Form Web App"
    type: "web"
    url: "https://httpbin.org/forms/post"
    timeout: 30
actions:
  - name: "navigate_to_form"
    type: "navigate"
    value: "https://httpbin.org/forms/post"
  - name: "wait_for_form"
    type: "wait"
    wait_time: 2
  - name: "fill_customer_name"
    type: "fill"
    selector: "input[name='custname']"
    value: "John Doe"
  - name: "fill_customer_email"
    type: "fill"
    selector: "input[name='custemail']"
    value: "john.doe@example.com"
  - name: "select_topping"
    type: "click"
    selector: "input[value='bacon']"
  - name: "submit_form"
    type: "submit"
    selector: "form"
  - name: "wait_response"
    type: "wait"
    wait_time: 3
  - name: "capture_result"
    type: "screenshot"
    parameters:
      filename: "form_result.png"
settings:
  enable_metrics: true
  log_level: "debug"
EOF

    # Multi-platform configuration
    cat > "$DATA_DIR/configs/multi_platform.yaml" << 'EOF'
name: "Multi-Platform Test"
output: "./output/multi_platform"
apps:
  - name: "Web Application"
    type: "web"
    url: "https://httpbin.org/html"
    timeout: 30
  - name: "Desktop Application"
    type: "desktop"
    path: "/Applications/Calculator.app"
    timeout: 15
  - name: "Mobile Application"
    type: "mobile"
    platform: "android"
    emulator: true
    device: "emulator-5554"
    timeout: 20
actions:
  - name: "test_web_app"
    type: "screenshot"
    parameters:
      filename: "web_test.png"
  - name: "test_desktop_app"
    type: "screenshot"
    parameters:
      filename: "desktop_test.png"
  - name: "test_mobile_app"
    type: "screenshot"
    parameters:
      filename: "mobile_test.png"
settings:
  screenshot_format: "png"
  quality: 85
  enable_metrics: true
EOF

    # Performance testing configuration
    cat > "$DATA_DIR/configs/performance.yaml" << 'EOF'
name: "Performance Test"
output: "./output/performance"
apps:
  - name: "Performance Test App"
    type: "web"
    url: "https://httpbin.org/delay/1"
    timeout: 30
actions:
  - name: "start_recording"
    type: "record"
    duration: 30
    parameters:
      filename: "performance_test.mp4"
  - name: "navigate_performance"
    type: "navigate"
    value: "https://httpbin.org/delay/1"
  - name: "wait_performance"
    type: "wait"
    wait_time: 2
  - name: "screenshot_performance"
    type: "screenshot"
    parameters:
      filename: "performance_test.png"
  - name: "fill_performance_form"
    type: "fill"
    selector: "input.test"
    value: "performance_test_data"
  - name: "click_performance"
    type: "click"
    selector: "button.test"
  - name: "wait_after_click"
    type: "wait"
    wait_time: 3
  - name: "final_screenshot"
    type: "screenshot"
    parameters:
      filename: "performance_final.png"
settings:
  screenshot_format: "png"
  video_format: "mp4"
  quality: 90
  enable_metrics: true
  window_width: 1920
  window_height: 1080
EOF

    # Error handling configuration
    cat > "$DATA_DIR/configs/error_handling.yaml" << 'EOF'
name: "Error Handling Test"
output: "./output/error_handling"
apps:
  - name: "Valid App"
    type: "web"
    url: "https://httpbin.org/html"
    timeout: 30
  - name: "Invalid URL App"
    type: "web"
    url: "https://non-existent-domain-12345.com"
    timeout: 10
  - name: "Invalid Desktop App"
    type: "desktop"
    path: "/non/existent/path/app"
    timeout: 10
actions:
  - name: "test_valid_app"
    type: "screenshot"
    parameters:
      filename: "valid_test.png"
  - name: "test_invalid_url"
    type: "navigate"
    value: "https://invalid-url-for-testing.com"
  - name: "test_invalid_desktop"
    type: "screenshot"
    parameters:
      filename: "invalid_desktop.png"
settings:
  enable_metrics: true
  log_level: "debug"
EOF

    print_success "Test configurations generated"
}

# Generate HTML fixtures
generate_fixtures() {
    print_status "Generating HTML fixtures..."
    
    # Basic HTML page
    cat > "$FIXTURES_DIR/web/basic_page.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Basic Test Page</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .button { padding: 10px 20px; background: #007cba; color: white; border: none; cursor: pointer; margin: 10px; }
        .input { padding: 8px; margin: 5px; border: 1px solid #ccc; width: 300px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Basic Test Page</h1>
        <p>This is a basic HTML page for testing purposes.</p>
        
        <button id="testButton" class="button">Test Button</button>
        <button id="anotherButton" class="button">Another Button</button>
        
        <div id="testContent" style="margin-top: 20px; padding: 10px; background: #f0f0f0;">
            <p>Content will appear here when buttons are clicked.</p>
        </div>
    </div>
    
    <script>
        document.getElementById('testButton').addEventListener('click', function() {
            document.getElementById('testContent').innerHTML = '<p>Test button was clicked!</p>';
        });
        
        document.getElementById('anotherButton').addEventListener('click', function() {
            document.getElementById('testContent').innerHTML = '<p>Another button was clicked!</p>';
        });
    </script>
</body>
</html>
EOF

    # Login form page
    cat > "$FIXTURES_DIR/web/login_form.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login Form</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f4f4f4; margin: 0; padding: 40px; }
        .login-container { max-width: 400px; margin: 0 auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h2 { text-align: center; color: #333; margin-bottom: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; color: #555; }
        .input { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        .checkbox { margin-right: 10px; }
        .button { width: 100%; padding: 12px; background: #007cba; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; }
        .button:hover { background: #005a87; }
        .error { color: red; font-size: 14px; margin-top: 10px; }
        .success { color: green; font-size: 14px; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="login-container">
        <h2>Login</h2>
        <form id="loginForm">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" class="input" required>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" class="input" required>
            </div>
            <div class="form-group">
                <label>
                    <input type="checkbox" id="remember" name="remember" class="checkbox">
                    Remember me
                </label>
            </div>
            <div class="form-group">
                <button type="submit" class="button">Login</button>
            </div>
            <div id="message"></div>
        </form>
    </div>
    
    <script>
        document.getElementById('loginForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            var username = document.getElementById('username').value;
            var password = document.getElementById('password').value;
            var remember = document.getElementById('remember').checked;
            var messageDiv = document.getElementById('message');
            
            if (username === 'admin' && password === 'password') {
                messageDiv.className = 'success';
                messageDiv.innerHTML = 'Login successful! Welcome, ' + username + '!';
            } else {
                messageDiv.className = 'error';
                messageDiv.innerHTML = 'Invalid username or password';
            }
        });
    </script>
</body>
</html>
EOF

    # Complex interactive page
    cat > "$FIXTURES_DIR/web/interactive_page.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Interactive Test Page</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f8f9fa; }
        .container { max-width: 1200px; margin: 0 auto; }
        .card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .tabs { display: flex; border-bottom: 1px solid #ddd; }
        .tab { padding: 10px 20px; cursor: pointer; border: 1px solid transparent; border-bottom: none; }
        .tab.active { background: white; border-color: #ddd; border-bottom: 1px solid white; margin-bottom: -1px; }
        .tab-content { display: none; padding: 20px 0; }
        .tab-content.active { display: block; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .item { padding: 15px; border: 1px solid #ddd; border-radius: 4px; }
        .progress-bar { width: 100%; height: 20px; background: #e9ecef; border-radius: 10px; overflow: hidden; }
        .progress-fill { height: 100%; background: linear-gradient(90deg, #007cba, #005a87); transition: width 0.3s; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <h1>Interactive Test Page</h1>
            <div class="tabs">
                <div class="tab active" data-tab="forms">Forms</div>
                <div class="tab" data-tab="dynamic">Dynamic Content</div>
                <div class="tab" data-tab="data">Data Display</div>
            </div>
            
            <div id="forms" class="tab-content active">
                <h3>Form Elements</h3>
                <div class="grid">
                    <div class="item">
                        <label for="text-input">Text Input:</label>
                        <input type="text" id="text-input" placeholder="Enter text..." class="input" style="width: 100%; padding: 8px; margin-top: 5px;">
                    </div>
                    <div class="item">
                        <label for="select-input">Select Dropdown:</label>
                        <select id="select-input" class="input" style="width: 100%; padding: 8px; margin-top: 5px;">
                            <option value="">Choose an option</option>
                            <option value="option1">Option 1</option>
                            <option value="option2">Option 2</option>
                            <option value="option3">Option 3</option>
                        </select>
                    </div>
                    <div class="item">
                        <label for="checkbox-group">Checkboxes:</label>
                        <div style="margin-top: 5px;">
                            <label><input type="checkbox" name="options" value="checkbox1"> Checkbox 1</label><br>
                            <label><input type="checkbox" name="options" value="checkbox2"> Checkbox 2</label><br>
                            <label><input type="checkbox" name="options" value="checkbox3"> Checkbox 3</label>
                        </div>
                    </div>
                    <div class="item">
                        <label for="radio-group">Radio Buttons:</label>
                        <div style="margin-top: 5px;">
                            <label><input type="radio" name="radio-options" value="radio1"> Radio 1</label><br>
                            <label><input type="radio" name="radio-options" value="radio2"> Radio 2</label><br>
                            <label><input type="radio" name="radio-options" value="radio3"> Radio 3</label>
                        </div>
                    </div>
                </div>
                <button id="submit-form" style="margin-top: 20px; padding: 10px 20px; background: #28a745; color: white; border: none; border-radius: 4px; cursor: pointer;">Submit Form</button>
                <div id="form-result" style="margin-top: 20px; padding: 10px; background: #e9ecef; border-radius: 4px; display: none;"></div>
            </div>
            
            <div id="dynamic" class="tab-content">
                <h3>Dynamic Content</h3>
                <div class="grid">
                    <div class="item">
                        <button id="add-item" style="padding: 8px 16px; background: #007cba; color: white; border: none; border-radius: 4px; cursor: pointer;">Add Item</button>
                        <button id="remove-item" style="margin-left: 10px; padding: 8px 16px; background: #dc3545; color: white; border: none; border-radius: 4px; cursor: pointer;">Remove Item</button>
                        <div id="dynamic-items" style="margin-top: 15px;">
                            <div class="item" style="background: #f8f9fa; padding: 10px; margin-bottom: 5px;">Item 1</div>
                        </div>
                    </div>
                    <div class="item">
                        <button id="toggle-content" style="padding: 8px 16px; background: #6c757d; color: white; border: none; border-radius: 4px; cursor: pointer;">Toggle Content</button>
                        <div id="toggle-content-area" style="margin-top: 15px; padding: 15px; background: #e3f2fd; border-radius: 4px; display: none;">
                            <h4>Hidden Content</h4>
                            <p>This content can be toggled on and off.</p>
                        </div>
                    </div>
                </div>
            </div>
            
            <div id="data" class="tab-content">
                <h3>Data Display</h3>
                <div class="item">
                    <h4>Progress Bar</h4>
                    <div class="progress-bar">
                        <div class="progress-fill" id="progress-fill" style="width: 0%;"></div>
                    </div>
                    <button id="start-progress" style="margin-top: 10px; padding: 8px 16px; background: #28a745; color: white; border: none; border-radius: 4px; cursor: pointer;">Start Progress</button>
                </div>
                <div class="item">
                    <h4>Data Table</h4>
                    <table style="width: 100%; border-collapse: collapse; margin-top: 10px;">
                        <thead>
                            <tr style="background: #f8f9fa;">
                                <th style="border: 1px solid #ddd; padding: 8px;">ID</th>
                                <th style="border: 1px solid #ddd; padding: 8px;">Name</th>
                                <th style="border: 1px solid #ddd; padding: 8px;">Status</th>
                                <th style="border: 1px solid #ddd; padding: 8px;">Action</th>
                            </tr>
                        </thead>
                        <tbody id="data-table-body">
                            <tr>
                                <td style="border: 1px solid #ddd; padding: 8px;">1</td>
                                <td style="border: 1px solid #ddd; padding: 8px;">Test Item 1</td>
                                <td style="border: 1px solid #ddd; padding: 8px;">Active</td>
                                <td style="border: 1px solid #ddd; padding: 8px;">
                                    <button class="action-btn" data-id="1" style="padding: 4px 8px; background: #007cba; color: white; border: none; border-radius: 2px; cursor: pointer;">Action</button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
    
    <script>
        // Tab switching
        document.querySelectorAll('.tab').forEach(tab => {
            tab.addEventListener('click', function() {
                document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
                document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
                
                this.classList.add('active');
                document.getElementById(this.dataset.tab).classList.add('active');
            });
        });
        
        // Dynamic content
        let itemCount = 1;
        document.getElementById('add-item').addEventListener('click', function() {
            itemCount++;
            const itemsDiv = document.getElementById('dynamic-items');
            const newItem = document.createElement('div');
            newItem.className = 'item';
            newItem.style.background = '#f8f9fa';
            newItem.style.padding = '10px';
            newItem.style.marginBottom = '5px';
            newItem.textContent = 'Item ' + itemCount;
            itemsDiv.appendChild(newItem);
        });
        
        document.getElementById('remove-item').addEventListener('click', function() {
            const itemsDiv = document.getElementById('dynamic-items');
            if (itemsDiv.children.length > 1) {
                itemsDiv.removeChild(itemsDiv.lastChild);
            }
        });
        
        document.getElementById('toggle-content').addEventListener('click', function() {
            const content = document.getElementById('toggle-content-area');
            content.style.display = content.style.display === 'none' ? 'block' : 'none';
        });
        
        // Progress bar
        document.getElementById('start-progress').addEventListener('click', function() {
            const progressFill = document.getElementById('progress-fill');
            let progress = 0;
            const interval = setInterval(() => {
                progress += 5;
                progressFill.style.width = progress + '%';
                if (progress >= 100) {
                    clearInterval(interval);
                }
            }, 100);
        });
        
        // Table actions
        document.querySelectorAll('.action-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const id = this.dataset.id;
                alert('Action clicked for item ID: ' + id);
            });
        });
    </script>
</body>
</html>
EOF

    print_success "HTML fixtures generated"
}

# Generate mock applications
generate_mocks() {
    print_status "Generating mock applications..."
    
    # Mock web server
    cat > "$MOCK_APPS_DIR/web/mock_server.py" << 'EOF'
#!/usr/bin/env python3
"""
Mock HTTP server for testing Panoptic web interactions
Provides various endpoints for testing different scenarios
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import time
import urllib.parse

class MockHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        """Handle GET requests"""
        if self.path == '/':
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            
            html = '''
<!DOCTYPE html>
<html>
<head><title>Mock Server</title></head>
<body>
    <h1>Mock Test Server</h1>
    <form id="testForm" method="POST" action="/submit">
        <input name="test_input" type="text" placeholder="Test Input">
        <button type="submit">Submit</button>
    </form>
    <button id="testButton">Test Button</button>
    <div id="result"></div>
</body>
</html>
'''
            self.wfile.write(html.encode())
            
        elif self.path == '/delay':
            delay = int(urllib.parse.parse_qs(urllib.parse.urlparse(self.path).query).get('delay', [2])[0])
            time.sleep(delay)
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            self.wfile.write(f'<html><body><h1>Delayed response after {delay} seconds</h1></body></html>'.encode())
            
        elif self.path == '/json':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            json_data = {'message': 'Mock JSON response', 'status': 'success'}
            self.wfile.write(json.dumps(json_data).encode())
            
        elif self.path == '/error':
            self.send_response(500)
            self.end_headers()
            self.wfile.write(b'Internal Server Error')
            
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b'Not Found')
    
    def do_POST(self):
        """Handle POST requests"""
        if self.path == '/submit':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            
            html = f'''
<!DOCTYPE html>
<html>
<head><title>Form Submitted</title></head>
<body>
    <h1>Form Submitted Successfully!</h1>
    <p>Data received: {post_data.decode()}</p>
    <a href="/">Back to form</a>
</body>
</html>
'''
            self.wfile.write(html.encode())
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b'Not Found')
    
    def log_message(self, format, *args):
        """Override to reduce log output"""
        pass

if __name__ == '__main__':
    server = HTTPServer(('localhost', 8080), MockHandler)
    print("Mock server running on http://localhost:8080")
    print("Available endpoints:")
    print("  GET  /         - Main test page")
    print("  GET  /delay    - Delayed response (?delay=seconds)")
    print("  GET  /json     - JSON response")
    print("  GET  /error    - Server error")
    print("  POST /submit   - Form submission")
    server.serve_forever()
EOF

    chmod +x "$MOCK_APPS_DIR/web/mock_server.py"
    
    # Mock mobile application (Android)
    cat > "$MOCK_APPS_DIR/mobile/MockActivity.java" << 'EOF'
package com.panoptic.test;

import android.app.Activity;
import android.os.Bundle;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import android.widget.Toast;

public class MockActivity extends Activity {
    private EditText inputText;
    private Button submitButton;
    private Button testButton;
    private TextView resultText;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_mock);

        inputText = findViewById(R.id.inputText);
        submitButton = findViewById(R.id.submitButton);
        testButton = findViewById(R.id.testButton);
        resultText = findViewById(R.id.resultText);

        submitButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                String input = inputText.getText().toString();
                resultText.setText("Submitted: " + input);
                Toast.makeText(MockActivity.this, "Form submitted!", Toast.LENGTH_SHORT).show();
            }
        });

        testButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                resultText.setText("Test button clicked!");
                Toast.makeText(MockActivity.this, "Test action!", Toast.LENGTH_SHORT).show();
            }
        });
    }
}
EOF

    # Mock mobile layout
    mkdir -p "$MOCK_APPS_DIR/mobile/res/layout"
    cat > "$MOCK_APPS_DIR/mobile/res/layout/activity_mock.xml" << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<LinearLayout xmlns:android="http://schemas.android.com/apk/res/android"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    android:orientation="vertical"
    android:padding="16dp">

    <TextView
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:text="Panoptic Test Application"
        android:textSize="20sp"
        android:textStyle="bold"
        android:layout_marginBottom="20dp"
        android:gravity="center" />

    <EditText
        android:id="@+id/inputText"
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:hint="Enter test input"
        android:padding="12dp"
        android:background="@android:drawable/edit_text"
        android:layout_marginBottom="16dp" />

    <Button
        android:id="@+id/submitButton"
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:text="Submit"
        android:layout_marginBottom="8dp" />

    <Button
        android:id="@+id/testButton"
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:text="Test Action"
        android:layout_marginBottom="16dp" />

    <TextView
        android:id="@+id/resultText"
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:text="Results will appear here"
        android:padding="12dp"
        android:background="#f0f0f0"
        android:gravity="center" />

</LinearLayout>
EOF

    print_success "Mock applications generated"
}

# Generate test templates
generate_templates() {
    print_status "Generating test templates..."
    
    # Configuration template
    cat > "$TEMPLATES_DIR/configs/test_template.yaml" << 'EOF'
name: "{{TEST_NAME}}"
output: "{{OUTPUT_DIR}}"
apps:
  - name: "{{APP_NAME}}"
    type: "{{APP_TYPE}}"
    {{#if APP_URL}}url: "{{APP_URL}}"{{/if}}
    {{#if APP_PATH}}path: "{{APP_PATH}}"{{/if}}
    {{#if APP_PLATFORM}}platform: "{{APP_PLATFORM}}"{{/if}}
    timeout: {{TIMEOUT}}
actions:
{{#each ACTIONS}}
  - name: "{{name}}"
    type: "{{type}}"
    {{#if selector}}selector: "{{selector}}"{{/if}}
    {{#if value}}value: "{{value}}"{{/if}}
    {{#if wait_time}}wait_time: {{wait_time}}{{/if}}
    {{#if duration}}duration: {{duration}}{{/if}}
    {{#if parameters}}parameters: {{parameters}}{{/if}}
{{/each}}
settings:
  screenshot_format: "{{SCREENSHOT_FORMAT}}"
  video_format: "{{VIDEO_FORMAT}}"
  quality: {{QUALITY}}
  enable_metrics: {{ENABLE_METRICS}}
  log_level: "{{LOG_LEVEL}}"
EOF

    # Test scenario template
    cat > "$TEMPLATES_DIR/scenarios/scenario_template.yaml" << 'EOF'
name: "{{SCENARIO_NAME}}"
description: "{{SCENARIO_DESCRIPTION}}"
prerequisites:
  - "{{PREREQ_1}}"
  - "{{PREREQ_2}}"
steps:
{{#each STEPS}}
  - name: "{{name}}"
    description: "{{description}}"
    action: "{{action}}"
    expected_result: "{{expected_result}}"
    timeout: {{timeout}}
    {{#if parameters}}parameters: {{parameters}}{{/if}}
{{/each}}
test_data:
  {{#each TEST_DATA}}
  {{@key}}: "{{this}}"
  {{/each}}
success_criteria:
  - "{{SUCCESS_CRITERION_1}}"
  - "{{SUCCESS_CRITERION_2}}"
EOF

    print_success "Test templates generated"
}

# Generate test images and assets
generate_assets() {
    print_status "Generating test assets..."
    
    # Create test images (using Python PIL if available)
    cat > "$FIXTURES_DIR/images/generate_test_images.py" << 'EOF'
#!/usr/bin/env python3
"""
Generate test images for Panoptic testing
"""

try:
    from PIL import Image, ImageDraw, ImageFont
    PIL_AVAILABLE = True
except ImportError:
    PIL_AVAILABLE = False

import os

def create_test_images():
    if not PIL_AVAILABLE:
        print("PIL not available, skipping image generation")
        return
    
    # Create images directory
    os.makedirs('test_images', exist_ok=True)
    
    # Create a simple test image
    img = Image.new('RGB', (800, 600), color='white')
    draw = ImageDraw.Draw(img)
    
    # Draw some test elements
    draw.rectangle([50, 50, 750, 550], outline='black', width=2)
    draw.rectangle([100, 100, 700, 500], fill='lightblue')
    
    # Add text
    try:
        font = ImageFont.truetype('/System/Library/Fonts/Arial.ttf', 40)
    except:
        font = ImageFont.load_default()
    
    draw.text((300, 250), 'Test Image', fill='black', font=font)
    draw.text((200, 320), 'Panoptic Testing', fill='darkblue', font=font)
    
    # Draw buttons
    draw.rectangle([250, 400, 350, 450], fill='green')
    draw.text((265, 415), 'Button', fill='white', font=font)
    
    draw.rectangle([400, 400, 500, 450], fill='red')
    draw.text((415, 415), 'Cancel', fill='white', font=font)
    
    # Save image
    img.save('test_images/test_screenshot.png')
    
    # Create a smaller thumbnail
    img.thumbnail((200, 150))
    img.save('test_images/test_thumbnail.png')
    
    print("Test images generated successfully")

if __name__ == '__main__':
    create_test_images()
EOF

    # Try to generate images
    cd "$FIXTURES_DIR/images"
    if command -v python3 &> /dev/null; then
        python3 generate_test_images.py 2>/dev/null || print_warning "Could not generate test images (PIL not available)"
    fi
    cd - > /dev/null
    
    # Create placeholder images if PIL not available
    if [ ! -f "$FIXTURES_DIR/images/test_images/test_screenshot.png" ]; then
        # Create simple SVG placeholder
        cat > "$FIXTURES_DIR/images/placeholder.svg" << 'EOF'
<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">
  <rect width="800" height="600" fill="white"/>
  <rect x="50" y="50" width="700" height="500" fill="none" stroke="black" stroke-width="2"/>
  <rect x="100" y="100" width="600" height="400" fill="#e6f3ff"/>
  <text x="400" y="300" text-anchor="middle" font-family="Arial" font-size="32" fill="black">
    Test Image Placeholder
  </text>
  <rect x="250" y="400" width="100" height="50" fill="green"/>
  <rect x="450" y="400" width="100" height="50" fill="red"/>
</svg>
EOF
        
        # Convert to PNG if ImageMagick is available
        if command -v convert &> /dev/null; then
            convert "$FIXTURES_DIR/images/placeholder.svg" "$FIXTURES_DIR/images/test_screenshot.png"
        fi
    fi
    
    print_success "Test assets generated"
}

# Generate test scenarios
generate_scenarios() {
    print_status "Generating test scenarios..."
    
    # E-commerce scenario
    cat > "$DATA_DIR/scenarios/ecommerce_flow.yaml" << 'EOF'
name: "E-commerce User Journey"
description: "Complete e-commerce flow from browsing to checkout"
prerequisites:
  - "E-commerce website is accessible"
  - "Test user account exists"
test_data:
  username: "testuser@example.com"
  password: "testpassword123"
  product_name: "Test Product"
steps:
  - name: "navigate_to_homepage"
    description: "Navigate to e-commerce homepage"
    action: "navigate"
    value: "https://example-ecommerce.com"
    expected_result: "Homepage loads successfully"
    timeout: 10
  - name: "search_product"
    description: "Search for test product"
    action: "fill"
    selector: "input[name='search']"
    value: "{{test_data.product_name}}"
    expected_result: "Search input is filled"
    timeout: 5
  - name: "submit_search"
    description: "Submit search form"
    action: "click"
    selector: "button[type='submit']"
    expected_result: "Search results are displayed"
    timeout: 10
  - name: "select_product"
    description: "Click on first product in results"
    action: "click"
    selector: ".product-item:first-child a"
    expected_result: "Product detail page loads"
    timeout: 10
  - name: "add_to_cart"
    description: "Add product to cart"
    action: "click"
    selector: "button[id='add-to-cart']"
    expected_result: "Product added to cart"
    timeout: 5
  - name: "navigate_to_cart"
    description: "Navigate to shopping cart"
    action: "navigate"
    value: "https://example-ecommerce.com/cart"
    expected_result: "Cart page loads with product"
    timeout: 10
  - name: "proceed_to_checkout"
    description: "Proceed to checkout"
    action: "click"
    selector: "button[id='checkout']"
    expected_result: "Checkout page loads"
    timeout: 10
  - name: "login"
    description: "Login with test credentials"
    action: "fill"
    selector: "input[name='email']"
    value: "{{test_data.username}}"
    expected_result: "Email field is filled"
    timeout: 5
  - name: "enter_password"
    description: "Enter password"
    action: "fill"
    selector: "input[name='password']"
    value: "{{test_data.password}}"
    expected_result: "Password field is filled"
    timeout: 5
  - name: "submit_login"
    description: "Submit login form"
    action: "click"
    selector: "button[type='submit']"
    expected_result: "User is logged in"
    timeout: 10
  - name: "complete_purchase"
    description: "Complete purchase"
    action: "click"
    selector: "button[id='complete-purchase']"
    expected_result: "Purchase is completed"
    timeout: 10
  - name: "capture_confirmation"
    description: "Capture order confirmation"
    action: "screenshot"
    parameters:
      filename: "order_confirmation.png"
    expected_result: "Screenshot is saved"
    timeout: 5
success_criteria:
  - "User successfully navigates through the entire e-commerce flow"
  - "Product is added to cart and purchased"
  - "Order confirmation is displayed"
  - "All steps complete within expected timeframes"
EOF

    # Form validation scenario
    cat > "$DATA_DIR/scenarios/form_validation.yaml" << 'EOF'
name: "Form Validation Testing"
description: "Test various form validation scenarios"
prerequisites:
  - "Form page is accessible"
test_data:
  valid_email: "valid@example.com"
  invalid_email: "invalid-email"
  short_password: "123"
  valid_password: "ValidPassword123!"
  long_text: "This is a very long text that exceeds normal field limits and should be truncated or rejected by the form validation system for being too long."
steps:
  - name: "navigate_to_form"
    description: "Navigate to registration form"
    action: "navigate"
    value: "https://example.com/register"
    expected_result: "Registration form loads"
    timeout: 10
  - name: "test_empty_fields"
    description: "Test form submission with empty fields"
    action: "submit"
    selector: "form[id='registration']"
    expected_result: "Validation errors shown"
    timeout: 5
  - name: "test_invalid_email"
    description: "Test with invalid email format"
    action: "fill"
    selector: "input[name='email']"
    value: "{{test_data.invalid_email}}"
    expected_result: "Email validation error"
    timeout: 5
  - name: "test_short_password"
    description: "Test with password too short"
    action: "fill"
    selector: "input[name='password']"
    value: "{{test_data.short_password}}"
    expected_result: "Password length error"
    timeout: 5
  - name: "test_valid_data"
    description: "Test with valid data"
    action: "fill"
    selector: "input[name='email']"
    value: "{{test_data.valid_email}}"
    expected_result: "Email field accepts valid input"
    timeout: 5
  - name: "test_valid_password"
    description: "Test with valid password"
    action: "fill"
    selector: "input[name='password']"
    value: "{{test_data.valid_password}}"
    expected_result: "Password field accepts valid input"
    timeout: 5
  - name: "submit_valid_form"
    description: "Submit form with valid data"
    action: "submit"
    selector: "form[id='registration']"
    expected_result: "Registration successful"
    timeout: 10
  - name: "capture_success_state"
    description: "Capture successful registration state"
    action: "screenshot"
    parameters:
      filename: "registration_success.png"
    expected_result: "Success screenshot saved"
    timeout: 5
success_criteria:
  - "Form properly validates empty fields"
  - "Email validation works correctly"
  - "Password validation works correctly"
  - "Valid data is accepted and processed"
  - "Success flow completes correctly"
EOF

    print_success "Test scenarios generated"
}

# Create test data index
create_index() {
    print_status "Creating test data index..."
    
    cat > "$DATA_DIR/README.md" << 'EOF'
# Panoptic Test Data

This directory contains comprehensive test data for Panoptic testing.

## Directory Structure

```
test_data/
├── configs/          # Test configuration files
├── fixtures/         # HTML and other test fixtures
├── scenarios/        # Complete test scenarios
└── assets/          # Test images and media

fixtures/
├── web/             # Web page HTML fixtures
├── mobile/           # Mobile app layouts
├── desktop/          # Desktop application mocks
└── images/          # Test images and media

mock_apps/
├── web/             # Mock web server
├── mobile/           # Mock mobile applications
└── desktop/          # Mock desktop apps

templates/
├── configs/          # Configuration templates
└── scenarios/        # Scenario templates
```

## Usage

### Test Configurations
Located in `configs/`:
- `web_basic.yaml` - Basic web application test
- `web_form.yaml` - Form interaction test
- `multi_platform.yaml` - Multi-platform testing
- `performance.yaml` - Performance testing
- `error_handling.yaml` - Error handling scenarios

### Test Fixtures
Located in `fixtures/`:
- `web/basic_page.html` - Basic HTML test page
- `web/login_form.html` - Login form with validation
- `web/interactive_page.html` - Complex interactive page

### Test Scenarios
Located in `scenarios/`:
- `ecommerce_flow.yaml` - Complete e-commerce user journey
- `form_validation.yaml` - Form validation testing

### Mock Applications
Located in `mock_apps/`:
- `web/mock_server.py` - Mock HTTP server for web testing
- `mobile/` - Android mock application files

## Quick Start

1. **Run basic web test:**
   ```bash
   ./panoptic run test_data/configs/web_basic.yaml
   ```

2. **Start mock server:**
   ```bash
   python3 test_data/mock_apps/web/mock_server.py
   ```

3. **Use test fixtures:**
   ```bash
   python3 -m http.server 8080 --directory test_data/fixtures/web/
   ```

4. **Run scenario tests:**
   ```bash
   ./panoptic run test_data/configs/web_form.yaml
   ```

## Data Categories

### Functional Test Data
- Basic page interactions
- Form submissions
- Navigation flows
- Click and fill operations

### Performance Test Data
- Large page loads
- Complex interactions
- Resource-intensive operations
- Timing measurements

### Error Test Data
- Invalid configurations
- Network failures
- Element not found scenarios
- Timeout conditions

### Security Test Data
- Injection attempts
- Cross-site scripting scenarios
- Authentication bypass attempts
- Data privacy tests

## Maintenance

- Update configurations regularly to match application changes
- Refresh test fixtures to maintain relevance
- Add new scenarios as features are added
- Review and update mock applications

## Contributing

When adding new test data:
1. Follow the directory structure
2. Use descriptive names
3. Include proper documentation
4. Test with actual Panoptic runs
5. Update this README
EOF

    print_success "Test data index created"
}

# Main execution
main() {
    print_status "Starting Panoptic test data generation..."
    print_status "========================================="
    
    init_directories
    
    case $GENERATE_TYPE in
        "configs")
            generate_configs
            ;;
        "fixtures")
            generate_fixtures
            ;;
        "mocks")
            generate_mocks
            ;;
        "templates")
            generate_templates
            ;;
        "scenarios")
            generate_scenarios
            ;;
        "assets")
            generate_assets
            ;;
        "all")
            generate_configs
            generate_fixtures
            generate_mocks
            generate_templates
            generate_scenarios
            generate_assets
            ;;
        *)
            print_error "Invalid type: $GENERATE_TYPE"
            exit 1
            ;;
    esac
    
    create_index
    
    print_status "========================================="
    print_success "Test data generation completed!"
    print_status "Generated data available in: $DATA_DIR/"
    print_status "Fixtures available in: $FIXTURES_DIR/"
    print_status "Mock apps available in: $MOCK_APPS_DIR/"
    print_status "Templates available in: $TEMPLATES_DIR/"
}

# Check dependencies
if ! command -v python3 &> /dev/null; then
    print_warning "Python3 not available - some features may be limited"
fi

if [ "$VERBOSE" = true ]; then
    set -x
fi

# Run main function
main "$@"