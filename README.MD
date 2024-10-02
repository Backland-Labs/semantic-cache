> [!Important]
> SemCache is under active development right now 🛠️
> 
> Give us a ⭐ to support the project and watch our repositories not to miss any update. Appreciate your interest 🙏

## SemCache - Semantic Caching for LLM Applications

SemCache is an open-source semantic cache microservice for LLM applications. SemCache is aimed at advanced LLM application developers looking for a secure, performant semantic cache for their applications.


## Why Use SemCache?
SemCache abstracts away the complexity of semantic cache by packing the configuration of the database and API routes into a simple endpoint. 

### When to use SemCache
SemCache fills the need for lightweight tools that provide core functionality without the bloat of the existing libraries. SemCache is built with cloud native deployments in mind.

## Endpoints

`/check` - Checks if there is any data in the cache - GET

`/write` - Upserts data in the cache - POST

## Quickstart

To begin:

`git clone https://github.com/Backland-Labs/semantic-cache.git`

Create a .env file in the repo root with the following contents:

```bash
QDRANT_HOST="qdrant"
QDRANT_COLLECTION="llmcache" # this can be anything
```

Check that the image for the cache container in the `docker-compose.yaml` file is correct:

`ghcr.io/backland-labs/semantic-cache/semantic-cache-a8044edf3c2796d6a5951c230da20a70:latest`

Then:

`docker-compose up`

The cache will be available on `http://localhost:8080/`