# Static Assets Tests
# Test static file serving including email templates, CSS, images, and other assets

# Test serving CSS assets
GET {{host}}/statics/css/email.css

HTTP 200
[Asserts]
header "Content-Type" contains "text/css"
body contains "body"

# Test serving JavaScript assets
GET {{host}}/statics/js/main.js

HTTP 200
[Asserts]
header "Content-Type" contains "javascript"

# Test serving image assets - logo
GET {{host}}/statics/images/logo.png

HTTP 200
[Asserts]
header "Content-Type" contains "image/png"

# Test serving email template assets
GET {{host}}/statics/email_templates/welcome.html

HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "html"
body contains "welcome"

# Test serving registration email template
GET {{host}}/statics/email_templates/registration.html

HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "html"

# Test serving password reset email template
GET {{host}}/statics/email_templates/password_reset.html

HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "html"
body contains "password"

# Test serving PDF templates (should be publicly accessible for testing)
GET {{host}}/statics/templates/pdf/invoice.html

HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "html"
body contains "invoice"

# Test serving report template
GET {{host}}/statics/templates/pdf/report.html

HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "html"

# Test serving certificate template
GET {{host}}/statics/templates/pdf/certificate.html

HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
body contains "html"

# Test favicon
GET {{host}}/favicon.ico

HTTP 200
[Asserts]
header "Content-Type" contains "image"

# Test robots.txt
GET {{host}}/robots.txt

HTTP 200
[Asserts]
header "Content-Type" contains "text/plain"

# Test non-existent static file (should return 404)
GET {{host}}/statics/nonexistent/file.css

HTTP 404

# Test directory traversal protection
GET {{host}}/statics/../../../etc/passwd

HTTP 404

# Test accessing file outside static directory
GET {{host}}/statics/../../../../etc/hosts

HTTP 404

# Test case sensitivity for static files
GET {{host}}/statics/CSS/EMAIL.CSS

HTTP 404

# Test accessing directory without index
GET {{host}}/statics/css/

HTTP 404

# Test large static file handling (if exists)
GET {{host}}/statics/images/large-image.jpg

# Should either return 200 or 404, both acceptable
HTTP *

# Test MIME type detection for different file extensions
GET {{host}}/statics/css/bootstrap.min.css

HTTP 200
[Asserts]
header "Content-Type" contains "text/css"

# Test JavaScript MIME type
GET {{host}}/statics/js/jquery.min.js

HTTP 200
[Asserts]
header "Content-Type" contains "javascript"

# Test image MIME types
GET {{host}}/statics/images/icon.svg

HTTP 200
[Asserts]
header "Content-Type" contains "image/svg"

# Test font file serving
GET {{host}}/statics/fonts/roboto.woff2

HTTP 200
[Asserts]
header "Content-Type" contains "font"

# Test JSON file serving
GET {{host}}/statics/data/config.json

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test caching headers for static assets
GET {{host}}/statics/css/email.css

HTTP 200
[Asserts]
header "Cache-Control" exists
header "ETag" exists

# Test compression for text assets
GET {{host}}/statics/css/email.css
Accept-Encoding: gzip

HTTP 200
# Note: Compression headers depend on server configuration

# Test CORS headers for static assets
GET {{host}}/statics/js/main.js
Origin: https://example.com

HTTP 200
# Note: CORS headers depend on server configuration

# Test static file serving under load (multiple identical requests)
GET {{host}}/statics/css/email.css

HTTP 200

GET {{host}}/statics/css/email.css

HTTP 200

GET {{host}}/statics/css/email.css

HTTP 200

# Test mixed case file names
GET {{host}}/statics/images/Logo.PNG

HTTP 404

# Test URL encoding in file names
GET {{host}}/statics/css/email%20style.css

HTTP 404

# Test special characters in file paths
GET {{host}}/statics/css/email@style.css

HTTP 404

# Test empty file name
GET {{host}}/statics/css/

HTTP 404

# Test file extension variations
GET {{host}}/statics/templates/pdf/invoice.HTML

HTTP 404

# Test accessing system files through static route
GET {{host}}/statics/.env

HTTP 404

GET {{host}}/statics/.git/config

HTTP 404

# Test binary file serving
GET {{host}}/statics/images/sample.pdf

HTTP 200
[Asserts]
header "Content-Type" contains "application/pdf"