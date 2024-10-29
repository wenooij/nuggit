CREATE TABLE
    IF NOT EXISTS Resources (
        ID INTEGER NOT NULL,
        APIVersion TEXT,
        Kind TEXT NOT NULL CHECK (Kind IN ('pipe', 'view', 'rule')),
        Version TEXT,
        Description TEXT,
        PipeID INTEGER,
        ViewID INTEGER,
        RuleID INTEGER,
        CHECK (
            COALESCE(PipeID, ViewID, RuleID) IS NOT NULL
            AND (
                PipeID IS NULL
                OR ViewID IS NULL
                OR RuleID IS NULL
            )
        ),
        UNIQUE (PipeID),
        UNIQUE (ViewID),
        UNIQUE (RuleID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        FOREIGN KEY (ViewID) REFERENCES Views (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        FOREIGN KEY (RuleID) REFERENCES Rules (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS ResourceLabels (
        ID INTEGER NOT NULL,
        ResourceID INTEGER,
        Label TEXT NOT NULL,
        UNIQUE (ResourceID, Label),
        FOREIGN KEY (ResourceID) REFERENCES Resources (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS Pipes (
        ID INTEGER NOT NULL,
        Name TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Digest TEXT NOT NULL CHECK (Digest GLOB '[0-9a-f][0-9a-f]*'),
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
    IF NOT EXISTS PipeDependencies (
        ID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        ReferencedID INTEGER NOT NULL CHECK (PipeID != ReferencedID),
        UNIQUE (PipeID, ReferencedID),
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        FOREIGN KEY (ReferencedID) REFERENCES Pipes (ID) ON UPDATE CASCADE ON DELETE CASCADE,
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
        FOREIGN KEY (ViewID) REFERENCES Views (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS PipesByView ON ViewPipes (PipeID);

CREATE TABLE
    IF NOT EXISTS Rules (
        ID INTEGER NOT NULL,
        Hostname TEXT,
        URLPattern TEXT,
        AlwaysTrigger BOOLEAN,
        Disable BOOLEAN,
        UNIQUE (Hostname, URLPattern, AlwaysTrigger, Disable),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS RuleLabels (
        ID INTEGER NOT NULL,
        RuleID INTEGER,
        Label TEXT NOT NULL,
        UNIQUE (RuleID, Label),
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS Events (
        ID INTEGER NOT NULL,
        PlanID INTEGER NOT NULL,
        Implicit BOOLEAN,
        URL TEXT,
        Timestamp TIMESTAMP,
        FOREIGN KEY (PlanID) REFERENCES Plans (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS EventsByPlan ON Events (PlanID);

CREATE TABLE
    IF NOT EXISTS Plans (
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
    IF NOT EXISTS PlanPipes (
        ID INTEGER NOT NULL,
        PlanID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        UNIQUE (PlanID, PipeID),
        FOREIGN KEY (PlanID) REFERENCES Plans (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE TABLE
    IF NOT EXISTS Results (
        ID INTEGER NOT NULL,
        EventID INTEGER NOT NULL,
        PipeID INTEGER NOT NULL,
        SequenceID INTEGER NOT NULL,
        TypeNumber INTEGER,
        Result BLOB,
        UNIQUE (EventID, PipeID, SequenceID),
        FOREIGN KEY (EventID) REFERENCES Events (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        FOREIGN KEY (PipeID) REFERENCES Pipes (ID) ON UPDATE CASCADE ON DELETE CASCADE,
        PRIMARY KEY (ID AUTOINCREMENT)
    );

CREATE INDEX IF NOT EXISTS ResultsByEvent ON Results (EventID);

CREATE INDEX IF NOT EXISTS ResultsByPipe ON Results (PipeID);