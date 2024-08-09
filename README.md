# idalloc
**Idalloc** is a scalable open-source service framework designed for generating unique IDs. It leverages MySQL, Redis, and in-memory caching to achieve high performance and efficiency.

**Features:**
- **Batch Generation:** Supports batch generation of incremental IDs.
- **High Performance:** Primarily allocates IDs in-memory, supplemented by asynchronous generation from Redis.
- **Reliability:** Asynchronous writes to the database ensure data persistence, with periodic synchronization to restore Redis from MySQL, minimizing the impact of potential Redis data loss.

**Note:** This service is highly dependent on Redis. In the event of a Redis outage, the service's availability may be affected. It is recommended to implement a fallback plan in the client to handle such scenarios.