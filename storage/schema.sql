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
        Name TEXT NOT NULL CHECK (Name GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Version TEXT NOT NULL CHECK (Version GLOB '[a-zA-Z][a-zA-Z0-9-]*'),
        Digest TEXT NOT NULL CHECK (Digest GLOB '[0-9a-f][0-9a-f]*'),
        FOREIGN KEY (Name) REFERENCES Pipes (Name),
        FOREIGN KEY (Digest) REFERENCES Pipes (Digest),
        PRIMARY KEY (Name, Version, Digest)
    );

CREATE TABLE
    IF NOT EXISTS Collections (
        CollectionID TEXT NOT NULL PRIMARY KEY,
        Name TEXT NOT NULL,
        AlwaysTrigger BOOLEAN,
        Hostname TEXT,
        URLPattern TEXT,
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        )
    );

CREATE INDEX IF NOT EXISTS CollectionsByName ON Collections (Name);

CREATE INDEX IF NOT EXISTS CollectionsByHostname ON Collections (Hostname);

CREATE TABLE
    IF NOT EXISTS CollectionPipes (
        CollectionID TEXT NOT NULL,
        PipeName TEXT NOT NULL,
        PipeDigest TEXT NOT NULL,
        FOREIGN KEY (CollectionID) REFERENCES Collections (CollectionID),
        FOREIGN KEY (PipeName) REFERENCES Pipes (PipeName),
        FOREIGN KEY (PipeDigest) REFERENCES Pipes (PipeDigest),
        PRIMARY KEY (CollectionID)
    );

CREATE TABLE
    IF NOT EXISTS CollectionData (
        ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        CollectionID TEXT NOT NULL,
        TriggerID TEXT,
        DataRow TEXT NOT NULL CHECK (
            json_valid (DataRow)
            AND json_type (DataRow) = 'array'
        ),
        FOREIGN KEY (TriggerID) REFERENCES Triggers (TriggerID)
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