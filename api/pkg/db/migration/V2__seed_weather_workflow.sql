-- 1. Seed the Node Library
INSERT INTO node_library (id, node_type, base_label, base_description, metadata)
VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'start', 'Start', 'Begin weather check workflow',
     '{"hasHandles": {"source": true, "target": false}}'),

    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'form', 'User Input', 'Process collected data - name, email, location',
     '{"hasHandles": {"source": true, "target": true}, "inputFields": ["name", "email", "city"], "outputVariables": ["name", "email", "city"]}'),

    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'integration', 'Weather API', 'Fetch current temperature for {{city}}',
     '{"hasHandles": {"source": true, "target": true}, "inputVariables": ["city"], "apiEndpoint": "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true", "options": [{"city": "Sydney", "lat": -33.8688, "lon": 151.2093}, {"city": "Melbourne", "lat": -37.8136, "lon": 144.9631}, {"city": "Brisbane", "lat": -27.4698, "lon": 153.0251}, {"city": "Perth", "lat": -31.9505, "lon": 115.8605}, {"city": "Adelaide", "lat": -34.9285, "lon": 138.6007}], "outputVariables": ["temperature"]}'),

    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'condition', 'Check Condition', 'Evaluate temperature threshold',
     '{"hasHandles": {"source": ["true", "false"], "target": true}, "conditionExpression": "temperature {{operator}} {{threshold}}", "outputVariables": ["conditionMet"]}'),

    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'email', 'Send Alert', 'Email weather alert notification',
     '{"hasHandles": {"source": true, "target": true}, "inputVariables": ["name", "city", "temperature"], "emailTemplate": {"subject": "Weather Alert", "body": "Weather alert for {{city}}! Temperature is {{temperature}}°C!"}, "outputVariables": ["emailSent"]}'),

    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 'end', 'Complete', 'Workflow execution finished',
     '{"hasHandles": {"source": false, "target": true}}');

-- 2. Create the Workflow Container
INSERT INTO workflows (id, name)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'Weather Check System');

-- 3. Place Nodes on the Canvas (Instances)
INSERT INTO workflow_node_instances (workflow_id, instance_id, node_library_id, x_pos, y_pos)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'start',       'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -160, 300),
    ('550e8400-e29b-41d4-a716-446655440000', 'form',        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 152, 304),
    ('550e8400-e29b-41d4-a716-446655440000', 'weather-api', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 460, 304),
    ('550e8400-e29b-41d4-a716-446655440000', 'condition',   'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 794, 304),
    ('550e8400-e29b-41d4-a716-446655440000', 'email',       'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 1096, 88),
    ('550e8400-e29b-41d4-a716-446655440000', 'end',         'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', 1360, 302);

-- 4. Create the Edges
INSERT INTO workflow_edges (workflow_id, edge_id, source_instance_id, target_instance_id, source_handle, animated, label, style_props, label_style)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'e1', 'start', 'form', null,
     true, 'Initialize', '{"stroke": "#10b981", "strokeWidth": 3}', null),

    ('550e8400-e29b-41d4-a716-446655440000', 'e2', 'form', 'weather-api', null,
     true, 'Submit Data', '{"stroke": "#3b82f6", "strokeWidth": 3}', null),

    ('550e8400-e29b-41d4-a716-446655440000', 'e3', 'weather-api', 'condition', null,
     true, 'Temperature Data', '{"stroke": "#f97316", "strokeWidth": 3}', null),

    ('550e8400-e29b-41d4-a716-446655440000', 'e4', 'condition', 'email', 'true',
     true, E'\u2713 Condition Met', '{"stroke": "#10b981", "strokeWidth": 3}', '{"fill": "#10b981", "fontWeight": "bold"}'),

    ('550e8400-e29b-41d4-a716-446655440000', 'e5', 'condition', 'end', 'false',
     true, E'\u2717 No Alert Needed', '{"stroke": "#6b7280", "strokeWidth": 3}', '{"fill": "#6b7280", "fontWeight": "bold"}'),

    ('550e8400-e29b-41d4-a716-446655440000', 'e6', 'email', 'end', null,
     true, 'Alert Sent', '{"stroke": "#ef4444", "strokeWidth": 2}', '{"fill": "#ef4444", "fontWeight": "bold"}');
