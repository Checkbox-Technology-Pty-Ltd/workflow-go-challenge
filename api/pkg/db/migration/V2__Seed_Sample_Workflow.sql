-- Seed data for the sample weather workflow
-- This matches the hardcoded JSON in workflow.go

-- 1. Insert the workflow
INSERT INTO workflows (id, name, description, version) VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    'Weather Alert Workflow',
    'A workflow that checks weather conditions and sends email alerts',
    1
);

-- 2. Insert the nodes
INSERT INTO nodes (workflow_id, node_id, node_type, label, description, x_pos, y_pos, metadata) VALUES
(
    '550e8400-e29b-41d4-a716-446655440000',
    'start',
    'start',
    'Start',
    'Begin weather check workflow',
    -160,
    300,
    '{"hasHandles": {"source": true, "target": false}}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'form',
    'form',
    'User Input',
    'Process collected data - name, email, location',
    152,
    304,
    '{"hasHandles": {"source": true, "target": true}, "inputFields": ["name", "email", "city"], "outputVariables": ["name", "email", "city"]}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'weather-api',
    'integration',
    'Weather API',
    'Fetch current temperature for {{city}}',
    460,
    304,
    '{"hasHandles": {"source": true, "target": true}, "inputVariables": ["city"], "apiEndpoint": "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true", "options": [{"city": "Sydney", "lat": -33.8688, "lon": 151.2093}, {"city": "Melbourne", "lat": -37.8136, "lon": 144.9631}, {"city": "Brisbane", "lat": -27.4698, "lon": 153.0251}, {"city": "Perth", "lat": -31.9505, "lon": 115.8605}, {"city": "Adelaide", "lat": -34.9285, "lon": 138.6007}], "outputVariables": ["temperature"]}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'condition',
    'condition',
    'Check Condition',
    'Evaluate temperature threshold',
    794,
    304,
    '{"hasHandles": {"source": ["true", "false"], "target": true}, "conditionExpression": "temperature {{operator}} {{threshold}}", "outputVariables": ["conditionMet"]}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'email',
    'email',
    'Send Alert',
    'Email weather alert notification',
    1096,
    88,
    '{"hasHandles": {"source": true, "target": true}, "inputVariables": ["name", "city", "temperature"], "emailTemplate": {"subject": "Weather Alert", "body": "Weather alert for {{city}}! Temperature is {{temperature}}°C!"}, "outputVariables": ["emailSent"]}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'end',
    'end',
    'Complete',
    'Workflow execution finished',
    1360,
    302,
    '{"hasHandles": {"source": false, "target": true}}'::jsonb
);

-- 3. Insert the edges
INSERT INTO edges (workflow_id, edge_id, source_id, target_id, source_handle, edge_props) VALUES
(
    '550e8400-e29b-41d4-a716-446655440000',
    'e1',
    'start',
    'form',
    NULL,
    '{"type": "smoothstep", "animated": true, "style": {"stroke": "#10b981", "strokeWidth": 3}, "label": "Initialize"}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'e2',
    'form',
    'weather-api',
    NULL,
    '{"type": "smoothstep", "animated": true, "style": {"stroke": "#3b82f6", "strokeWidth": 3}, "label": "Submit Data"}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'e3',
    'weather-api',
    'condition',
    NULL,
    '{"type": "smoothstep", "animated": true, "style": {"stroke": "#f97316", "strokeWidth": 3}, "label": "Temperature Data"}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'e4',
    'condition',
    'email',
    'true',
    '{"type": "smoothstep", "animated": true, "style": {"stroke": "#10b981", "strokeWidth": 3}, "label": "✓ Condition Met", "labelStyle": {"fill": "#10b981", "fontWeight": "bold"}}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'e5',
    'condition',
    'end',
    'false',
    '{"type": "smoothstep", "animated": true, "style": {"stroke": "#6b7280", "strokeWidth": 3}, "label": "✗ No Alert Needed", "labelStyle": {"fill": "#6b7280", "fontWeight": "bold"}}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'e6',
    'email',
    'end',
    NULL,
    '{"type": "smoothstep", "animated": true, "style": {"stroke": "#ef4444", "strokeWidth": 2}, "label": "Alert Sent", "labelStyle": {"fill": "#ef4444", "fontWeight": "bold"}}'::jsonb
);
