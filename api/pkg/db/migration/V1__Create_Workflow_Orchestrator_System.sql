-- 1. Create the Workflows Table
-- This is the "Blueprint" header.
CREATE TABLE workflows (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create the Nodes Table
-- We pull out the "Minimum Structure" into columns for indexing and searchability.
CREATE TABLE nodes (
    workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE,
    node_id VARCHAR(100) NOT NULL, -- e.g., 'start', 'weather-api'
    node_type VARCHAR(50) NOT NULL, -- e.g., 'integration', 'condition'
    
    -- Minimum Structure (Stable Fields)
    label VARCHAR(255),
    description TEXT,
    x_pos FLOAT NOT NULL,
    y_pos FLOAT NOT NULL,
    
    -- The JSONB Compromise
    -- Stores type-specific data (apiEndpoint, emailTemplate, conditionExpression)
    metadata JSONB DEFAULT '{}'::jsonb,
    
    PRIMARY KEY (workflow_id, node_id)
);

-- 3. Create the Edges Table
-- We ensure that an edge cannot exist without a source and target in the SAME workflow.
CREATE TABLE edges (
    workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE,
    edge_id VARCHAR(100) NOT NULL,
    source_id VARCHAR(100) NOT NULL,
    target_id VARCHAR(100) NOT NULL,
    source_handle VARCHAR(50), -- Necessary for branching logic (true/false)
    
    -- Store styling and UI-specific edge data
    edge_props JSONB DEFAULT '{}'::jsonb,
    
    PRIMARY KEY (workflow_id, edge_id),
    -- The Composite Foreign Key ensures the nodes belong to the specific workflow
    FOREIGN KEY (workflow_id, source_id) REFERENCES nodes(workflow_id, node_id),
    FOREIGN KEY (workflow_id, target_id) REFERENCES nodes(workflow_id, node_id)
);

-- 4. Create the Execution Log Table
-- This stores the "Run Result" from your second JSON example.
CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL, -- 'completed', 'failed'
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- The final state of variables (e.g., city: Sydney, temp: 28.5)
    final_context JSONB DEFAULT '{}'::jsonb,
    
    -- The full step-by-step trace of the execution
    execution_trace JSONB NOT NULL 
);

-- 5. Indexes for Query Performance
-- Makes HandleGetWorkflow and History lookups near-instant
CREATE INDEX idx_nodes_by_workflow ON nodes(workflow_id);
CREATE INDEX idx_edges_by_workflow ON edges(workflow_id);
CREATE INDEX idx_executions_by_workflow ON workflow_executions(workflow_id);
