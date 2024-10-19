CREATE TABLE
    IF NOT EXISTS Pipes (
        Name TEXT NOT NULL CHECK (Name LIKE '[a-zA-Z_][a-zA-Z0-9_]*'),
        Digest TEXT NOT NULL CHECK (Digest LIKE '[0-9a-fA-F][0-9a-fA-F]*'),
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
        Name TEXT NOT NULL CHECK (Name LIKE '[a-zA-Z_][a-zA-Z0-9_]*'),
        Version TEXT NOT NULL CHECK (Version LIKE '[a-zA-Z_][a-zA-Z0-9_]*'),
        Digest TEXT NOT NULL CHECK (Digest LIKE '[0-9a-fA-F][0-9a-fA-F]*'),
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

CREATE TABLE
    IF NOT EXISTS CollectionData (
        ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        CollectionID TEXT NOT NULL,
        TriggerID TEXT,
        DataRow TEXT NOT NULL CHECK (
            json_valid (DataRow)
            AND json_type (DataRow) = 'array'
        )
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
        PipeID TEXT NOT NULL,
        Committed BOOLEAN,
        Results TEXT CHECK (
            Results is NULL
            OR json_valid (Results)
        )
    );