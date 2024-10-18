CREATE TABLE
    IF NOT EXISTS Pipes (
        PipeID TEXT NOT NULL PRIMARY KEY,
        Name TEXT,
        Spec TEXT CHECK (
            Spec IS NULL
            OR (
                json_valid (Spec)
                AND json_type (Spec) = 'object'
            )
        )
    );

CREATE TABLE
    IF NOT EXISTS Collections (
        CollectionID TEXT NOT NULL PRIMARY KEY,
        Name TEXT NOT NULL,
        AlwaysTrigger BOOLEAN,
        Hostname TEXT,
        URLPattern TEXT,
        Spec TEXT CHECK (
            State IS NULL
            OR (
                json_valid (State)
                AND json_type (State) = 'object'
            )
        ),
        Conditions TEXT CHECK (
            Conditions IS NULL
            OR (
                json_valid (Conditions)
                AND json_type (Conditions) = 'object'
            )
        )
    );

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
        Metadata TEXT CHECK (
            Metadata IS NULL
            OR (
                json_valid (Metadata)
                AND json_type (Metadata) = 'object'
            )
        )
    );

CREATE TABLE
    IF NOT EXISTS TriggerResults (
        ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        TriggerID TEXT NOT NULL,
        PipeID TEXT NOT NULL,
        Results TEXT CHECK (
            Results is NULL
            OR json_valid (Results)
        )
    );