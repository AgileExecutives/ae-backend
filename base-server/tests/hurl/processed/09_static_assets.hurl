# Static JSON API Tests 
# Tests static JSON file serving from the statics/json directory

# Test 1: Serve bundeslaender JSON data
GET http://localhost:8080/static/bundeslaender

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "Baden-Württemberg"
body contains "states"

# Test 2: Serve enumerations JSON data
GET http://localhost:8080/static/enumerations

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "focus_areas"

# Test 3: Serve diagnostic_std JSON data
GET http://localhost:8080/static/diagnostic_std

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test 4: Serve ferien-feiertage JSON data
GET http://localhost:8080/static/ferien-feiertage

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test 5: Serve jugendaemter JSON data
GET http://localhost:8080/static/jugendaemter

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test 6: Serve potential_customers JSON data
GET http://localhost:8080/static/potential_customers

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test 7: Test directory traversal protection - URL normalization makes this a 404
GET http://localhost:8080/static/../../../etc/passwd

HTTP 404

# Test 8: Test filename with dots - should be blocked by our validation
GET http://localhost:8080/static/test..test

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test 9: Test non-existent JSON file
GET http://localhost:8080/static/nonexistent_1761600534_af78c817

HTTP 404
[Asserts]
body contains "File not found"

# Test 10: Test filename with slash - router returns 404 
GET http://localhost:8080/static/test/path

HTTP 404

# Test 11: Test valid JSON response structure for bundeslaender
GET http://localhost:8080/static/bundeslaender

HTTP 200
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$.states.BW" == "Baden-Württemberg"
jsonpath "$.states.BY" == "Bayern"

# Test 12: Test valid JSON response structure for enumerations
GET http://localhost:8080/static/enumerations

HTTP 200
[Asserts]
header "Content-Type" == "application/json" 
jsonpath "$.focus_areas" exists
jsonpath "$.service_offers" exists