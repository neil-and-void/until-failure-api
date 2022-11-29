# Workout Logger API
<img width="1440" alt="Screen Shot 2022-11-13 at 11 47 42 AM" src="https://user-images.githubusercontent.com/46465568/201538851-99b051a7-b084-4919-993c-93f7efffd447.png">


The GraphQL api for a very simple workout logger app I built for myself to track progression of weight I'm lifting for certain exercises.

# Techonologies Used

- Go
- GORM
- GQLGen
- GraphQL
- PostgreSQL

# Prereqs

- A postgres db url
- Go installed on your machine

# Setup

1. Clone repo
2. `cd` into the root of the repo
3. Have copy contents of `.test.env` into a new `.env` file
4. Fill in and replace secrets and postgres database connection parameters 
5. Run `make dev` to start test env

# Commands

- `make dev`: start dev environment
- `make test`: run all test files
- `make format`: format all code within repo
- `make regenerate`: regenerate graphql resolvers from `schema.graphqls`
