CREATE TABLE
    IF NOT EXISTS Pipes (
        Name TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Digest TEXT NOT NULL CHECK (Digest GLOB '[0-9a-f][0-9a-f]*'),
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        ),
        PRIMARY KEY (Name, Digest)
    );

CREATE INDEX IF NOT EXISTS PipesByNameDigest ON Pipes (CONCAT (Name, '@', Digest));

CREATE TABLE
    IF NOT EXISTS PipeVersions (
        Name TEXT NOT NULL,
        Version TEXT NOT NULL,
        Digest TEXT NOT NULL CHECK (Digest GLOB '[0-9a-f][0-9a-f]*'),
        FOREIGN KEY (Name) REFERENCES Pipes (Name),
        FOREIGN KEY (Digest) REFERENCES Pipes (Digest),
        PRIMARY KEY (Name, Version, Digest)
    );

CREATE TABLE
    IF NOT EXISTS Collections (
        Name TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Digest TEXT NOT NULL CHECK (Digest GLOB '[0-9a-f][0-9a-f]*'),
        AlwaysTrigger BOOLEAN,
        Hostname TEXT,
        URLPattern TEXT,
        Disabled BOOLEAN,
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        ),
        PRIMARY KEY (Name, Digest)
    );

CREATE INDEX IF NOT EXISTS CollectionsByNameDigest ON Collections (CONCAT (Name, '@', Digest));

CREATE INDEX IF NOT EXISTS CollectionsByHostname ON Collections (Hostname);

CREATE TABLE
    IF NOT EXISTS CollectionPipes (
        CollectionName TEXT NOT NULL,
        CollectionDigest TEXT NOT NULL,
        PipeName TEXT NOT NULL,
        PipeDigest TEXT NOT NULL,
        FOREIGN KEY (CollectionName) REFERENCES Collections (Name),
        FOREIGN KEY (CollectionDigest) REFERENCES Collections (Digest),
        FOREIGN KEY (PipeName) REFERENCES Pipes (Name),
        FOREIGN KEY (PipeDigest) REFERENCES Pipes (Digest),
        PRIMARY KEY (CollectionName, CollectionDigest, PipeName, PipeDigest)
    );

CREATE TABLE
    IF NOT EXISTS Triggers (
        TriggerID TEXT NOT NULL PRIMARY KEY,
        Committed BOOLEAN,
        Plan TEXT CHECK (
            Plan IS NULL
            OR (
                json_valid (Plan)
                AND json_type (Plan) = 'object'
            )
        ),
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        )
    );

CREATE TABLE
    IF NOT EXISTS TriggerResults (
        ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        TriggerID TEXT NOT NULL,
        PipeName TEXT NOT NULL,
        PipeDigest TEXT NOT NULL,
        Committed BOOLEAN,
        Results TEXT CHECK (
            Results is NULL
            OR json_valid (Results)
        ),
        FOREIGN KEY (TriggerID) REFERENCES Triggers (TriggerID),
        FOREIGN KEY (PipeName) REFERENCES Pipes (PipeName),
        FOREIGN KEY (PipeDigest) REFERENCES Pipes (PipeDigest)
    );