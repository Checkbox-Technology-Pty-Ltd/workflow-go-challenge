-- 1. Global Node Library (Reusable Blueprints)
CREATE TABLE node_library (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_type VARCHAR(50) NOT NULL,
    base_label VARCHAR(255),
    base_description TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Audit & Lifecycle
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE -- Soft delete field
);

-- 2. Workflows (The Container)
CREATE TABLE workflows (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,

    -- Audit & Lifecycle
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 3. Node Instances (Mapping Library Nodes to a Canvas)
CREATE TABLE workflow_node_instances (
    workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE,
    instance_id VARCHAR(100) NOT NULL, -- 'start', 'weather-api', etc.
    node_library_id UUID REFERENCES node_library(id),

    x_pos FLOAT NOT NULL,
    y_pos FLOAT NOT NULL,

    -- Audit & Lifecycle
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- No deleted_at here; usually we hard-delete instances when removed from a specific canvas

    PRIMARY KEY (workflow_id, instance_id)
);

-- 4. Workflow Edges (The Connections)
CREATE TABLE workflow_edges (
    workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE,
    edge_id VARCHAR(100) NOT NULL,
    source_instance_id VARCHAR(100) NOT NULL,
    target_instance_id VARCHAR(100) NOT NULL,
    source_handle VARCHAR(50),

    edge_type VARCHAR(50) DEFAULT 'smoothstep',
    animated BOOLEAN DEFAULT true,
    label VARCHAR(255),
    style_props JSONB DEFAULT '{}'::jsonb,
    label_style JSONB,

    -- Audit & Lifecycle
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (workflow_id, edge_id),
    FOREIGN KEY (workflow_id, source_instance_id)
        REFERENCES workflow_node_instances(workflow_id, instance_id),
    FOREIGN KEY (workflow_id, target_instance_id)
        REFERENCES workflow_node_instances(workflow_id, instance_id)
);

-- 5. Indexes for query performance
CREATE INDEX idx_node_library_active ON node_library(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflow_node_instances_library ON workflow_node_instances(node_library_id);
CREATE INDEX idx_workflow_edges_workflow ON workflow_edges(workflow_id);

-- 6. Automating modified_at (Trigger Function)
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to keep modified_at updated automatically
CREATE TRIGGER update_node_library_modtime BEFORE UPDATE ON node_library FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_workflows_modtime BEFORE UPDATE ON workflows FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_node_instances_modtime BEFORE UPDATE ON workflow_node_instances FOR EACH ROW EXECUTE FUNCTION update_modified_column();
CREATE TRIGGER update_edges_modtime BEFORE UPDATE ON workflow_edges FOR EACH ROW EXECUTE FUNCTION update_modified_column();
