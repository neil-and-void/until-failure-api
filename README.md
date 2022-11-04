# Workout Logger API

The GraphQL api for a very simple workout logger app I built for myself to track progression of weight I'm lifting for certain exercises.

# Techonologies Used

- Go
- GORM
- GQLGen
- GraphQL
- PostgreSQL

# Prereqs

- A postgres db url added to your `.env` file
- Go installed on your machine

# Setup

1. Clone repo
2. `cd` into the root of the repo
3. Have copy contents of `.test.env` into a new `.env` file
4. Run `make dev` to start test env

# Commands

- `make dev`: start dev environment
- `make test`: run all test files
- `make format`: format all code within repo
- `make regenerate`: regenerate graphql resolvers from `schema.graphqls`
