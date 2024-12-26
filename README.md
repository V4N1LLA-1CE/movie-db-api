# Movie API

A RESTful API for managing movie data with role-based access control.

## Tech Stack

- Go
- PostgreSQL
- Server-side sessions for authentication (stored in DB)
- Mailtrap with email templating for account notifications

## Features

- Permission-based access control (movies:read movies:write etc)
- Secure session-based authentication (No JWTs)
- Account activation after registration using Mailtrap with email templates
- Rate limiting for API endpoints
- Full CRUD operations for movies (if permissions allow)
- Secure password hashing

## License

MIT License - see LICENSE file for details
