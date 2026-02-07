-- Extend the node library with SMS and Flood API node types.
-- No schema changes required — the existing table structure supports
-- new node types through data alone.

INSERT INTO node_library (id, node_type, base_label, base_description, metadata)
VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', 'sms', 'Send SMS', 'Send an SMS notification',
     '{"hasHandles": {"source": true, "target": true}, "inputVariables": ["name", "phone", "message"], "outputVariables": ["deliveryStatus", "smsSent"]}'),

    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', 'flood', 'Flood Risk API', 'Check flood risk for a given location',
     '{"hasHandles": {"source": true, "target": true}, "inputVariables": ["city"], "apiEndpoint": "https://flood-api.open-meteo.com/v1/flood", "options": [{"city": "Sydney", "lat": -33.8688, "lon": 151.2093}, {"city": "Melbourne", "lat": -37.8136, "lon": 144.9631}, {"city": "Brisbane", "lat": -27.4698, "lon": 153.0251}], "outputVariables": ["floodRisk", "discharge"]}');
