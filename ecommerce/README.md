# E-commerce Service

## Project Structure
```
ecommerce/
├── cmd/
│   └── api/
│       └── main.go           # entrypoint for API server
├── internal/                 # private code for this service
│   ├── server/
│   │   └── router.go         # mux router & middleware
│   ├── db/
│   │   └── mongo.go          # mongo connect helper & getters
│   ├── handlers/             # http handlers (products, auth, orders, ...)
│   ├── models/               # domain models (Product, User, Order, ...)
│   ├── repository/           # DB access layer (interfaces + implementations)
│   └── middleware/           # auth, logging, recover, rate-limit
├── pkg/                      # reusable packages (if any)
├── configs/                  # configuration, schema, env examples
├── deployments/
│   └── docker-compose.yml
├── .env.example
├── Makefile
└── README.md
```
