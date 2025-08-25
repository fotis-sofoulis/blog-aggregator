# Blog Aggregator CLI (Gator)
A command-line tool to scrape and browse your favorite blog feeds, powered by PostgreSQL and Go.


## 1️⃣  Prerequisites
Make sure you have:
* Go installed (1.24+ recommended)
* PostgreSQL installed and running

## ⚙ Setup Instructions

1. Clone the repository:
```bash
git clone https://github.com/fotis-sofoulis/blog-aggregator.git && cd blog-aggregator/
```

2. Make sure you have Go 1.24+:
```bash
go version
```

3. Install dependencies:
```bash
go mod tidy
```

4. Build the binary:
```bash
go build -o gator
```

### 🛠 Config
5. Create a `.gatorconfig.json` file in your home directory with the following structure:
```json
{
  "db_url": "postgres://username:password@localhost:5432/database?sslmode=disable"
}
```

### 🛢Database Setup
6. Install `Goose` if you don’t have it:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

7. Move from the root folder to `sql/schema` and run migrations:
```bash
cd sql/schema && goose postgres "postgres://username:password@localhost:5432/gator" up
```

## ⌨️ Commands Overview
Here’s a list of the commands available in the Gator CLI. Use them to manage your account, follow feeds, and browse posts.
| Command       | Description |
|---------------|-------------|
| `login`       | Log in as an existing user. Requires the username. |
| `register`    | Create a new user account. Requires a username. |
| `reset`       | Reset the users table, deleting all users. |
| `users`       | List all registered users, highlighting the current user. |
| `agg`         | Continuously scrape all feeds at a specified interval (e.g., `1m` or `1h`). |
| `addfeed`     | Add a new feed and automatically follow it. Requires feed name and URL. |
| `feeds`       | List all feeds in the system along with the creator. |
| `follow`      | Follow a feed by its URL. |
| `following`   | Show all feeds you are currently following. |
| `unfollow`    | Unfollow a feed by its URL. |
| `browse`      | Browse posts from feeds you follow. Optional argument: number of posts to display (default 2). |

> Tip: You must be logged in to use commands that require authentication (`addfeed`, `follow`, `following`, `unfollow`, `browse`).

Command usage example:
```bash
./gator reset
```
