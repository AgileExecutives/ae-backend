# Static Assets Tests for Unburdy Server
# Test the limited static file serving that unburdy provides

# Test serving JSON data files
GET {{host}}/static/diagnostic_std

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "icd10"
body contains "kategorie"

# Test serving bundeslaender data
GET {{host}}/static/bundeslaender

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "name"

# Test serving enumerations data
GET {{host}}/static/enumerations

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving ferien-feiertage data
GET {{host}}/static/ferien-feiertage

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving jugendaemter data
GET {{host}}/static/jugendaemter

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving potential_customers data
GET {{host}}/static/potential_customers

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test listing available static JSON files
GET {{host}}/static

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "diagnostic_std.json"
body contains "bundeslaender.json"

# Test favicon serving
GET {{host}}/favicon.ico

HTTP 200
[Asserts]
header "Content-Type" contains "image"

# Test non-existent JSON file (should return 404)
GET {{host}}/static/nonexistent_file

HTTP 404

# Test case sensitivity for JSON files
GET {{host}}/static/DIAGNOSTIC_STD

HTTP 404

# Test accessing file without extension (should work)
GET {{host}}/static/diagnostic_std

HTTP 200

# Test accessing with wrong extension
GET {{host}}/static/diagnostic_std.xml

HTTP 404

# Test empty filename
GET {{host}}/static/

HTTP 404

# Test special characters in filename
GET {{host}}/static/diagnostic@std

HTTP 404

# Test directory traversal protection
GET {{host}}/static/../config

HTTP 404

# Test accessing system files through static route
GET {{host}}/static/.env

HTTP 404

GET {{host}}/static/..%2F..%2F.env

HTTP 404

# Test multiple requests to same file (caching behavior)
GET {{host}}/static/diagnostic_std

HTTP 200

GET {{host}}/static/diagnostic_std

HTTP 200

# Test file with underscores and hyphens
GET {{host}}/static/ferien-feiertage

HTTP 200