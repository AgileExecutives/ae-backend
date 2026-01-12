#!/bin/bash
cd /Users/alex/src/ae/backend/base-server
PGPASSWORD=password psql -h localhost -p 5432 -U postgres -d ae_saas_basic_test -c "
SELECT 'Templates:' as info, COUNT(*) as count FROM templates WHERE tenant_id = 1
UNION ALL
SELECT 'Tenants:', COUNT(*) FROM tenants
UNION ALL  
SELECT 'Template Contracts:', COUNT(*) FROM template_contracts;

SELECT 'Template Details:' as section, '' as details
UNION ALL
SELECT 'ID: ' || id::text, 'Channel: ' || channel || ', Key: ' || template_key || ', Type: ' || template_type 
FROM templates WHERE tenant_id = 1 LIMIT 5;
"