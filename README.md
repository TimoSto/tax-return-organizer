# tax-return-organizer

>This repo is still WIP

This repository hold services to help you organize documents for your tax return. It is used by two personas:

1. The collector: defines which documents and information a client must provide
2. The client: uploads the files and enters necessary data

### Personas & tenancy

This is a multi-tenant (SaaS-style) system: there can be many Collectors (e.g. different tax advisories), each managing their own set of Clients. A Collector is a company: several of its employees work in the same tenant. Clients log in to their Collector's tenant as themselves (one login per Client) and only see their own data.

## Usage by Collector

### Usecase *"Setup a Collector account"*:

- Give a name
- Give email address

### Usecase *"Create Client Personas"*:

- Client Personas represent the documents and information a client has to provide, if they do things, that are tax relevant
- Examples would be *Employed*, *Letting*, *Self-employed* or *ETF investing*
- The collector defines which documents and data (in form of defined structural data) the Personas need to provide and if the data needs to be provided monthly, quarterly, semi-annualy or annualy
- Some Client Personas should come preconfigured as templates

### Usecase *"Create a Client"*:

- Create a client object with a title and an email address

### Usecase *"Create Client Tax Year"*:

- Select a Client
- Enter a year (e.g. 2021)
- Assign Personas to this Client
    - The tax year keeps a reference to its Personas only for display and for updating the list (see below). Changing a Persona afterwards never changes an existing tax year
- Generate List of documents and data the Client needs to provide
    - For documents metadata can be specified, which the client needs to fill out
    - For data an optional free-text unit (e.g. *EUR*, *days*, *km*), a type (text, number, date or yes/no) and whether it is required can be specified (validation rules come later, see "Not in the first MVP")
- If a Persona changes, add the possibility to update this list explicitly, but no necessity (not possible on a closed tax year)

### Usecase *"Close Client Tax Year"*:

- The Collector can only close the tax year once every item is satisfied; while items are outstanding, closing is rejected
- After this the client cannot alter the files/data anymore

## Usage by Client

### Usecase *"View tax year"*:

- The Client needs to see all files and data they need to provide for a tax year
- They should not see the *Persona* layer, only the documents and data
- They need to see which items are still open and which are satisfied

### Usecase *"Upload files"*:

- For each item in the list a Client must be able to upload one or multiple files
- After upload the files must be rendered in the UI so the Client can check them
- Clients must be able to delete files
- Clients must be able to download files
- Only after uplaod Clients must be able to set and edit documents metadata

### Usecase *"Enter data"*:

- For standalone data required by the Persona, the Client must be able to enter it
- Required data must be entered before the item counts as satisfied

**Not in the first MVP:** 

- deadline tracking and upcoming/overdue notifications. The first MVP only tracks whether an item is satisfied or not; due dates and notifications are deferred to a later iteration.
- Authentication
- Validation rules on data and metadata values (e.g. *>0*). Planned as a `Rules` list on `ValueSpec`, each rule a named, parameterised built-in (functions cannot be persisted) that validates a typed value and returns an error. Type checking via `ValueType` already covers basic validation.

## Domain model

The domain types, how they relate and the design notes behind them are in [backend/docs/DOMAIN_MODEL.md](backend/docs/DOMAIN_MODEL.md).

## Architecture

A Go backend (REST API) and a SvelteKit BFF, backed by PostgreSQL, shipped as three containers in docker compose. Only the backend talks to Postgres. Documents (as blobs), their metadata and data entries are all stored in Postgres, not on the filesystem, so the Postgres container needs a named volume at `/var/lib/postgresql/data`; without it, `docker compose down` wipes all data.

The backend's internal structure (hexagonal layering, packages, design decisions) is described in [backend/docs/ARCHITECTURE.md](backend/docs/ARCHITECTURE.md).
