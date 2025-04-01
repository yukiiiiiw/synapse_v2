-- Create schedule schema
CREATE SCHEMA IF NOT EXISTS schedule;
set serach_path to schedule;

-- Resource PV and SKU per node
CREATE TABLE schedule.agent_node_resource (
    id BIGSERIAL PRIMARY KEY,
    status INTEGER NOT NULL,
    resource_type INTEGER NOT NULL,
    location VARCHAR(255) NOT NULL,
    region VARCHAR(255) NOT NULL,
    gpu_type VARCHAR(255),
    driver_version VARCHAR(255),
    random_type varchar(255) NULL,
    storage_type varchar(255) NULL,
    cloud_type varchar(255) NULL,
    max_gpu_count INTEGER DEFAULT 0,
    max_vcpu INTEGER DEFAULT 0,
    max_vram INTEGER DEFAULT 0,
    max_ram INTEGER DEFAULT 0,
    max_persistent_volume INTEGER NOT NULL,
    network_upload NUMERIC(38,18) DEFAULT 0.0,
    network_download NUMERIC(38,18) DEFAULT 0.0,
    disk_read_speed NUMERIC(38,18) DEFAULT 0.0,
    disk_write_speed NUMERIC(38,18) DEFAULT 0.0,
    random_type VARCHAR(255),
    storage_type VARCHAR(255),
    cloud_type VARCHAR(255),
    resource_md5 TEXT NOT NULL UNIQUE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.agent_node_resource IS 'Resource PV and SKU per node';
COMMENT ON COLUMN schedule.agent_node_resource.status IS '0:unavailable 1:available';
COMMENT ON COLUMN schedule.agent_node_resource.resource_type IS 'GPU, CPU';
COMMENT ON COLUMN schedule.agent_node_resource.gpu_type IS 'G4090';
COMMENT ON COLUMN schedule.agent_node_resource.max_gpu_count IS 'Node GPU count';
COMMENT ON COLUMN schedule.agent_node_resource.max_vcpu IS 'Node video CPU';
COMMENT ON COLUMN schedule.agent_node_resource.max_vram IS 'GB video RAM';
COMMENT ON COLUMN schedule.agent_node_resource.max_ram IS 'GB RAM';
COMMENT ON COLUMN schedule.agent_node_resource.max_persistent_volume IS 'GB node volume';
COMMENT ON COLUMN schedule.agent_node_resource.resource_md5 IS 'md5(resource_type:gpu_type:max_gpu_count:...:disk_price)';
COMMENT ON COLUMN schedule.agent_node_resource.version IS 'Optimistic Locking';

-- All nodes
CREATE TABLE schedule.agent_node_info (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    agent_node_id VARCHAR(255) NOT NULL,
    node_status INTEGER NOT NULL,
    status INTEGER NOT NULL,
    resource_type INTEGER NOT NULL,
    metric_info JSONB NOT NULL,
    resource_free_info JSONB NOT NULL,
    base_image VARCHAR(255),
    location VARCHAR(255) NOT NULL,
    region VARCHAR(255) NOT NULL,
    gpu_type VARCHAR(255) NOT NULL,
    total_gpu_count INTEGER DEFAULT 0,
    total_vram INTEGER DEFAULT 0,
    total_ram INTEGER DEFAULT 0,
    total_vcpu INTEGER DEFAULT 0,
    node_resource_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    last_heart_beat BIGINT,
    version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.agent_node_info IS 'All nodes';
COMMENT ON COLUMN schedule.agent_node_info.agent_id IS 'K8s cluster ID';
COMMENT ON COLUMN schedule.agent_node_info.agent_node_id IS 'K8s cluster node ID';
COMMENT ON COLUMN schedule.agent_node_info.node_status IS 'K8s cluster status';
COMMENT ON COLUMN schedule.agent_node_info.status IS 'Available status -1:offline 0:available 1:busy';
COMMENT ON COLUMN schedule.agent_node_info.resource_type IS 'GPU, CPU';
COMMENT ON COLUMN schedule.agent_node_info.base_image IS 'Linux v0.0.0, Unix v0.0.0';
COMMENT ON COLUMN schedule.agent_node_info.total_gpu_count IS 'Agent total GPU count';
COMMENT ON COLUMN schedule.agent_node_info.total_vram IS 'GB';
COMMENT ON COLUMN schedule.agent_node_info.total_ram IS 'GB';
COMMENT ON COLUMN schedule.agent_node_info.total_vcpu IS 'GB';
COMMENT ON COLUMN schedule.agent_node_info.version IS 'Optimistic Locking';

-- Agent
CREATE TABLE schedule.agent_info (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    agent_status INTEGER NOT NULL,
    status INTEGER NOT NULL,
    cloud_type VARCHAR(255) NOT NULL,
    metric_info JSONB NOT NULL,
    resource_free_info JSONB NOT NULL,
    base_image VARCHAR(255),
    location VARCHAR(255) NOT NULL,
    region VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    last_heart_beat BIGINT,
    version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.agent_info IS 'Agent';
COMMENT ON COLUMN schedule.agent_info.agent_id IS 'K8s cluster ID';
COMMENT ON COLUMN schedule.agent_info.agent_status IS 'K8s cluster status';
COMMENT ON COLUMN schedule.agent_info.status IS 'Available status -1:offline 0:available 1:busy';
COMMENT ON COLUMN schedule.agent_info.cloud_type IS 'Secure or community';
COMMENT ON COLUMN schedule.agent_info.base_image IS 'Linux v0.0.0, Unix v0.0.0';
COMMENT ON COLUMN schedule.agent_info.version IS 'Optimistic Locking';

CREATE INDEX idx_metric_info ON schedule.agent_info USING gin (metric_info);
CREATE INDEX idx_resource_free_info ON schedule.agent_info USING gin (resource_free_info);
CREATE UNIQUE INDEX udx_agent_id ON schedule.agent_info USING btree (agent_id);

-- Service
CREATE TABLE schedule.agent_service_info (
    id bigint NOT NULL primary key,
    saas_pod_id varchar(255) NOT NULL, -- saas podId as k8s service name
    
    agent_id varchar(255) NULL, -- k8s clusterId, strategy hit asynced
    service_name varchar(255) NULL, -- k8s service name
    service_info jsonb NULL, -- k8s service config (image,ip,port,gpu,cpu,disk...)
    metric_info jsonb NULL, -- pod realtime metric info (nodePort ingress pv http/tcp)
    
    status integer NOT NULL, -- -1:failure 0:deploying 10:updating 11:deleting 12:pausing 20:running 21:deleted 22:paused 
    created_at bigint NOT NULL,
    updated_at bigint NULL,
    service_updated_at bigint NULL,
    last_heart_beat bigint NULL,
    version int NOT NULL -- Optimistic Locking
);

CREATE UNIQUE INDEX udx_saas_pod_agent ON schedule.agent_service_info USING btree (saas_pod_id, agent_id);
CREATE INDEX idx_agent_id ON schedule.agent_service_info USING btree (agent_id);

-- Schedule service log
CREATE TABLE IF NOT EXISTS schedule.service_operate_log (
    id bigint NOT NULL primary key,
    service_info_id bigint NOT NULL,
    saas_pod_id varchar(255) NOT NULL,
    schedule_info jsonb NOT NULL,  -- saas required resources
    
    operate_type integer NOT NULL, -- operator type 0:deploy 1:update 2:delete 3:pause 4:resume 5:restart
    status integer NOT NULL, -- operator status 0:init 1:completed
    created_at bigint NOT NULL,
    updated_at bigint NULL,
    completed_at bigint NULL,
    version int NOT NULL -- Optimistic Locking
);

CREATE INDEX IF NOT EXISTS idx_service_operate_log_service_info_id ON schedule.service_operate_log(service_info_id);
CREATE INDEX IF NOT EXISTS idx_service_operate_log_saas_pod_id ON schedule.service_operate_log(saas_pod_id);


-- Strategy
CREATE TABLE schedule.route_rule (
    id BIGSERIAL PRIMARY KEY,
    rule_code VARCHAR(100) UNIQUE,
    description VARCHAR(255),
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.route_rule IS 'Strategy';
COMMENT ON COLUMN schedule.route_rule.rule_code IS 'config_match, resource_score';
COMMENT ON COLUMN schedule.route_rule.version IS 'Optimistic Locking';

CREATE TABLE schedule.route_chain (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(100) UNIQUE,
    resource_type INTEGER NOT NULL,
    rules JSONB,
    enabled BOOLEAN DEFAULT true,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.route_chain IS 'Strategy';
COMMENT ON COLUMN schedule.route_chain.resource_type IS 'GPU, CPU';
COMMENT ON COLUMN schedule.route_chain.rules IS '["config_match", "health_score"]';
COMMENT ON COLUMN schedule.route_chain.version IS 'Optimistic Locking';