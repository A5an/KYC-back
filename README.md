# KYC Backend

This guide offers a practical walkthrough on how to effectively set up and run the KYC API backend system.

## Migration
Upon each server start-up, the backend server is designed to automatically complete database migrations. This feature guarantees that the database schema is consistently up-to-date with the application.

## API DOC
Comprehensive API documentation can be accessed using Swagger UI at `http://localhost:<port>/swagger-ui/`.

## Configuring and Running Server in Development Mode

Setting up and running the server with development settings requires the following steps:

1. Set the configuration environment variable to 'dev':
   ```bash
   export CONFIG_NAME=dev
   ```
2. Start the main.go file found within the cmd directory:
    ```bash
    go run cmd/main.go
   ```
These steps will configure and start the server in dev mode.
Please remember to set up the project and its dependencies before attempting to start the server.