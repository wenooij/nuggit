CREATE TABLE
    IF NOT EXISTS Resources (
        ID INTEGER NOT NULL,
        APIVersion TEXT,
        Kind TEXT NOT NULL CHECK (Kind IN ('pipe', 'view')),
        Version TEXT,
        Description TEXT,
        PipeID INTEGER,
        ViewID INTEGER,
        CHECK (
            COALESCE(PipeID, ViewID) IS NOT NULL
            AND (
                PipeID IS NULL
                OR ViewID IS NULL
            )
        ),
        UNIQUE (PipeID),
        UNIQUE (ViewID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        FOREIGN KEY (ViewID) REFERENCES Views (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS ResourceLabels (
        ID INTEGER NOT NULL,
        ResourceID INTEGER,
        Label TEXT NOT NULL,
        UNIQUE (ResourceID, Label),
        FOREIGN KEY (ResourceID) REFERENCES Resources (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS Pipes (
        ID INTEGER NOT NULL,
        Name TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Digest TEXT NOT NULL CHECK (Digest GLOB '[0-9a-f][0-9a-f]*'),
        Disabled BOOLEAN,
        AlwaysTrigger BOOLEAN,
        TypeNumber INTEGER,
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
    IF NOT EXISTS PipeRules (
        ID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        RuleID INTEGER NOT NULL,
        UNIQUE (PipeID, RuleID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        FOREIGN KEY (RuleID) REFERENCES TriggerRules (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS PipeDependencies (
        ID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        ReferencedID INTEGER NOT NULL CHECK (PipeID != ReferencedID),
        UNIQUE (PipeID, ReferencedID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        FOREIGN KEY (ReferencedID) REFERENCES Pipes (ID),
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
    IF NOT EXISTS TriggerRules (
        ID INTEGER NOT NULL,
        Hostname TEXT NOT NULL CHECK (Hostname != ''),
        URLPattern TEXT,
        UNIQUE (Hostname, URLPattern),
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
        SequenceID INTEGER NOT NULL,
        TypeNumber INTEGER,
        Result BLOB,
        UNIQUE (EventID, PipeID, SequenceID),
        FOREIGN KEY (EventID) REFERENCES TriggerEvents (ID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS TriggerResultsByEvent ON TriggerResults (EventID);

CREATE INDEX IF NOT EXISTS TriggerResultsByPipe ON TriggerResults (PipeID);