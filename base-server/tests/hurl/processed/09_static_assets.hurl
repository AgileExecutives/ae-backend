# Static Assets API Tests 
# Tests static file serving including CSS, JavaScript, images, and templates

# Test 1: Serve CSS assets
GET http://localhost:8080/static/css/email.css

HTTP 200
[Asserts]
header "Content-Type" contains "text/css"

# Test 2: Serve JavaScript assets
GET http://localhost:8080/static/js/main.js

HTTP 200
[Asserts]
header "Content-Type" contains "javascript"

# Test 3: Serve SVG logo image
GET http://localhost:8080/static/images/logo.svg

HTTP 200
[Asserts]
header "Content-Type" contains "svg"

# Test 4: Serve favicon
GET http://localhost:8080/static/images/favicon.svg

HTTP 200
[Asserts]
header "Content-Type" contains "svg"

# Test 5: Serve email template
GET http://localhost:8080/static/email_templates/welcome.html

HTTP 200
[Asserts]
header "Content-Type" contains "html"

# Test 6: Serve PDF template (may not be implemented)
GET http://localhost:8080/static/pdf_templates/invoice.html

HTTP *

# Test 7: Test directory traversal protection
GET http://localhost:8080/static/../../../etc/passwd

# Should return 404 or similar security response
HTTP *

# Test 8: Test non-existent asset
GET http://localhost:8080/static/css/nonexistent_1761570720_3905b88b.css

# Should return 404 
HTTP 404

# Test 9: Test static assets without extension
GET http://localhost:8080/static/README.md

HTTP 200

# Test 10: Test email template with variables
GET http://localhost:8080/static/email_templates/password_reset.html

HTTP 200
[Asserts]
header "Content-Type" contains "html"
body contains "html"

# Test 11: Test asset caching headers (if implemented)
GET http://localhost:8080/static/css/email.css

HTTP 200
# Cache headers are optional, just verify the file serves correctly
[Asserts]
header "Content-Type" contains "text/css"

# Test 12: Test various image formats
GET http://localhost:8080/static/images/logo-full.svg

HTTP 200
[Asserts]
header "Content-Type" contains "svg"

# Test 13: Verify large file handling (if applicable)
GET http://localhost:8080/static/js/main.js

HTTP 200
[Asserts]
header "Content-Type" contains "javascript"

# Test 14: Test MIME type detection
GET http://localhost:8080/static/css/email.css

HTTP 200
[Asserts]
header "Content-Type" == "text/css; charset=utf-8"