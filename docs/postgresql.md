# PostgreSQL

PostgreSQL is an advanced, open-source object-relational database management system (ORDBMS) with an emphasis on extensibility and standards compliance. It supports both SQL (relational) and JSON (non-relational) querying.

## History

PostgreSQL was originally developed at the University of California, Berkeley, starting in 1986 as the POSTGRES project. It was renamed to PostgreSQL in 1996 to emphasize its SQL compliance.

## Features

### ACID Compliance
PostgreSQL fully supports ACID (Atomicity, Consistency, Isolation, Durability) properties, ensuring reliable transaction processing.

### Data Types
- Standard SQL types (INTEGER, VARCHAR, DATE, TIMESTAMP)
- Advanced types: JSON/JSONB, arrays, hstore (key-value), range types
- Geometric types and spatial data (PostGIS extension)
- Custom types via CREATE TYPE

### Indexing
- B-tree, Hash, GiST, SP-GiST, GIN, and BRIN index types
- Partial indexes for selective indexing
- Expression indexes for computed values
- Multi-column indexes

### Concurrency
- MVCC (Multi-Version Concurrency Control) for high-performance concurrent access
- Various locking mechanisms (row-level, table-level, advisory locks)
- Serializable snapshot isolation level

### Extensibility
- Custom functions in multiple languages (PL/pgSQL, PL/Python, PL/Perl)
- Extensions system (PostGIS, pg_trgm, uuid-ossp)
- Custom operators and type classes

## Architecture

### Process Model
PostgreSQL uses a client-server model with multiple processes:
- **Postmaster**: Parent process handling connections
- **Backend processes**: One per client connection
- **Background workers**: For maintenance tasks (WAL writer, checkpoint, autovacuum)
- **WAL (Write-Ahead Logging)**: Ensures data durability and enables point-in-time recovery

### Storage
- Tablespace management for physical storage layout
- TOAST (The Oversized-Attribute Storage Technique) for large field storage
- MVCC stores multiple versions of each row

## Common Extensions

- **PostGIS**: Geographic/spatial data support
- **pg_trgm**: Trigram-based text similarity search
- **pg_stat_statements**: Query performance tracking
- **TimescaleDB**: Time-series data optimization
- **Citus**: Distributed PostgreSQL for horizontal scaling
- **pgvector**: Vector similarity search for AI/ML

## Use Cases

- Web applications and APIs
- Data warehousing and analytics
- Geographic information systems (GIS)
- Financial systems requiring strict ACID compliance
- Scientific data management
- JSON/document storage with relational queries
- Real-time analytics and reporting
