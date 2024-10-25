CREATE TABLE
    IF NOT EXISTS Resources (
        ID INTEGER NOT NULL,
        Name TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Digest TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        APIVersion TEXT,
        Kind TEXT NOT NULL CHECK (Kind IN ('pipe', 'view')),
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        ),
        UNIQUE (Name, Digest),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS ResourceMetadata (
        ID INTEGER NOT NULL,
        ResourceID INTEGER NOT NULL,
        FOREIGN KEY (ResourceID) REFERENCES Resources (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS Pipes (
        ID INTEGER NOT NULL,
        Name TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Digest TEXT NOT NULL CHECK (Digest GLOB '[0-9a-f][0-9a-f]*'),
        CriteriaID INTEGER NOT NULL,
        TypeNumber INTEGER,
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        ),
        UNIQUE (Name, Digest),
        FOREIGN KEY (CriteriaID) REFERENCES TriggerCriteria (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX PipesByTriggerCriteria ON Pipes (CriteriaID);

CREATE TABLE
    IF NOT EXISTS PipeVersions (
        ID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        Version TEXT NOT NULL,
        UNIQUE (Version),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS PipesByVersion ON PipeVersions (PipeID, Version);

CREATE TABLE
    IF NOT EXISTS PipeDependencies (
        ID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        ReferencedID INTEGER NOT NULL CHECK (PipeID != ReferencedID),
        UNIQUE (PipeID, ReferencedID),
        FOREIGN KEY (PipeID, ReferencedID) REFERENCES Pipes (ID, ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS Views (
        ID INTEGER NOT NULL,
        UUID TEXT NOT NULL CHECK (UUID GLOB '????????-????-????-????-????????????'),
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        ),
        UNIQUE (UUID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS ViewPipes (
        ID INTEGER NOT NULL,
        ViewID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        UNIQUE (ViewID, PipeID),
        FOREIGN KEY (ViewID) REFERENCES Views (ID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS PipesByView ON ViewPipes (PipeID);

CREATE TABLE
    IF NOT EXISTS TriggerCriteria (
        ID INTEGER NOT NULL,
        Disabled BOOLEAN,
        AlwaysTrigger BOOLEAN,
        Hostname TEXT,
        URLPattern TEXT,
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS TriggerEvents (
        ID INTEGER NOT NULL,
        PlanID INTEGER NOT NULL,
        Implicit BOOLEAN,
        URL TEXT,
        Timestamp TIMESTAMP,
        FOREIGN KEY (PlanID) REFERENCES TriggerPlans (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS TriggerEventsByPlan ON TriggerEvents (PlanID);

CREATE TABLE
    IF NOT EXISTS TriggerPlans (
        ID INTEGER NOT NULL,
        UUID TEXT NOT NULL CHECK (UUID GLOB '????????-????-????-????-????????????'),
        Finished BOOLEAN,
        Plan TEXT CHECK (
            Plan IS NULL
            OR (
                json_valid (Plan)
                AND json_type (Plan) = 'object'
            )
        ),
        UNIQUE (UUID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS TriggerPlanPipes (
        ID INTEGER NOT NULL,
        PlanID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        UNIQUE (PlanID, PipeID),
        FOREIGN KEY (PlanID) REFERENCES TriggerPlans (ID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS TriggerResults (
        ID INTEGER NOT NULL,
        EventID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        TypeNumber INTEGER,
        Result BLOB,
        UNIQUE (EventID, BatchID),
        UNIQUE (BatchID, PipeID),
        UNIQUE (EventID, BatchID, PipeID),
        FOREIGN KEY (EventID) REFERENCES TriggerEvents (ID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS TriggerResultsByPipe ON TriggerResults (PipeID);