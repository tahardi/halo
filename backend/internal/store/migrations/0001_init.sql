CREATE TABLE attacks (
    id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 1 AND 64),
    started_at TEXT NOT NULL,
    ended_at TEXT,
    timezone TEXT NOT NULL CHECK (length(timezone) > 0),
    peak_severity INTEGER CHECK (peak_severity BETWEEN 0 AND 10),
    symptoms TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(symptoms)),
    symptoms_other TEXT,
    notes TEXT,
    lat REAL CHECK (lat BETWEEN -90 AND 90),
    lon REAL CHECK (lon BETWEEN -180 AND 180),
    source TEXT NOT NULL CHECK (source IN ('app', 'import')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT,
    CHECK (ended_at IS NULL OR ended_at >= started_at),
    CHECK ((lat IS NULL) = (lon IS NULL))
);

CREATE INDEX attacks_started_at ON attacks (started_at);

CREATE TABLE medications (
    id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 1 AND 64),
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    usual_amount REAL NOT NULL CHECK (usual_amount > 0),
    unit TEXT NOT NULL CHECK (length(trim(unit)) > 0),
    kind TEXT NOT NULL CHECK (kind IN ('acute', 'preventive')),
    archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1))
);

CREATE UNIQUE INDEX medications_name ON medications (name COLLATE NOCASE);

CREATE TABLE doses (
    id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 1 AND 64),
    attack_id TEXT NOT NULL REFERENCES attacks (id),
    medication_id TEXT NOT NULL REFERENCES medications (id),
    amount REAL NOT NULL CHECK (amount > 0),
    taken_at TEXT NOT NULL,
    relief TEXT CHECK (relief IN ('none', 'some', 'full')),
    deleted_at TEXT
);

CREATE INDEX doses_attack ON doses (attack_id);

CREATE TABLE recordings (
    id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 1 AND 64),
    attack_id TEXT NOT NULL REFERENCES attacks (id),
    audio_path TEXT NOT NULL,
    recorded_at TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'transcribed', 'extracted', 'confirmed', 'failed')),
    error TEXT,
    transcript TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX recordings_attack ON recordings (attack_id);

CREATE TABLE extractions (
    id INTEGER PRIMARY KEY,
    recording_id TEXT NOT NULL REFERENCES recordings (id),
    model TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    result TEXT NOT NULL CHECK (json_valid(result)),
    created_at TEXT NOT NULL
);

CREATE INDEX extractions_recording ON extractions (recording_id, id);

CREATE TABLE weather_hours (
    lat_round REAL NOT NULL,
    lon_round REAL NOT NULL,
    hour TEXT NOT NULL,
    pressure_hpa REAL NOT NULL,
    temperature_c REAL NOT NULL,
    humidity_pct REAL NOT NULL,
    precipitation_mm REAL NOT NULL,
    source TEXT NOT NULL,
    PRIMARY KEY (lat_round, lon_round, hour)
) WITHOUT ROWID;
