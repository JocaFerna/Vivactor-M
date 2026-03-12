curl -X POST http://localhost:8080/api/v1/report \
-H "Content-Type: application/json" \
-d '{
    "pathToCompiledMicroservices": "/home/joca/Thesis/repositories/downloads/piggymetrics/",
    "organizationPath": "com.piggymetrics",
    "outputPath": "metrics_report.wip"
}'