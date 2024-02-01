---
title: Environment Variables
weight: 300
---
Blinkfile uses the following environment variables for configuration:

| Variable                         | Description                                                     | Default |
|----------------------------------|-----------------------------------------------------------------|---------|
| ADMIN_USERNAME                   | The username for the admin user (leave blank for no admin user) |         |
| ADMIN_PASSWORD                   | The password for the admin user                                 |         |
| PORT                             | The port to listen on                                           | 8000    |
| DATA_DIR                         | The directory to store persistent data like file uploads        | /data   |
| RATE_LIMIT_UNAUTHENTICATED       | The rate limit per second for unauthenticated requests          | 2       |
| RATE_LIMIT_BURST_UNAUTHENTICATED | The burst rate limit per second for unauthenticated requests    | 5       |
| ENABLE_TEST_AUTOMATION           | Enable test automation endpoints                                | false   |
| EXPIRE_CHECK_CYCLE_TIME          | The time between clean up of expired files                      | 15m     |