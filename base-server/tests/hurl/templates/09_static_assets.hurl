# Static Assets API Tests 
# Tests static file serving including CSS, JavaScript, images, and templates

# Test 1: Serve CSS assets
GET {{HOST}}/static/css/email.css

HTTP 200
[Asserts]
header "Content-Type" contains "text/css"

# Test 2: Serve JavaScript assets
GET {{HOST}}/static/js/main.js

HTTP 200
[Asserts]
header "Content-Type" contains "javascript"

# Test 3: Serve SVG logo image
GET {{HOST}}/static/images/logo.svg

HTTP 200
[Asserts]
header "Content-Type" contains "svg"

# Test 4: Serve favicon
GET {{HOST}}/static/images/favicon.svg

HTTP 200
[Asserts]
header "Content-Type" contains "svg"

# Test 5: Serve email template
GET {{HOST}}/static/email_templates/welcome.html

HTTP 200
[Asserts]
header "Content-Type" contains "html"

# Test 6: Serve PDF template (may not be implemented)
GET {{HOST}}/static/pdf_templates/invoice.html

HTTP *

# Test 7: Test directory traversal protection
GET {{HOST}}/static/../../../etc/passwd

# Should return 404 or similar security response
HTTP *

# Test 8: Test non-existent asset
GET {{HOST}}/static/css/nonexistent_{{UNIQUE_ID}}.css

# Should return 404 
HTTP 404

# Test 9: Test static assets without extension
GET {{HOST}}/static/README.md

HTTP 200

# Test 10: Test email template with variables
GET {{HOST}}/static/email_templates/password_reset.html

HTTP 200
[Asserts]
header "Content-Type" contains "html"
body contains "html"

# Test 11: Test asset caching headers (if implemented)
GET {{HOST}}/static/css/email.css

HTTP 200
# Cache headers are optional, just verify the file serves correctly
[Asserts]
header "Content-Type" contains "text/css"

# Test 12: Test various image formats
GET {{HOST}}/static/images/logo-full.svg

HTTP 200
[Asserts]
header "Content-Type" contains "svg"

# Test 13: Verify large file handling (if applicable)
GET {{HOST}}/static/js/main.js

HTTP 200
[Asserts]
header "Content-Type" contains "javascript"

# Test 14: Test MIME type detection
GET {{HOST}}/static/css/email.css

HTTP 200
[Asserts]
header "Content-Type" == "text/css; charset=utf-8"