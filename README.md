# Book Rental System

Welcome to the Book Rental System! This project is designed to manage a library of books, allowing users to rent and return books. The system provides a set of APIs for book management, rental tracking, and more.

## Getting Started

### Prerequisites

- [Go (Golang)](https://golang.org/dl/)
- [MongoDB](https://www.mongodb.com/try/download/community)

### Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/book-rental-system.git
    cd book-rental-system
    ```

2. Install dependencies:

    ```bash
    go mod download
    ```

3. Run the application:

    ```bash
    go run main.go
    ```

The server will be accessible at `http://localhost:9000`.

## API Endpoints

- **Add Book:** `POST /books`
- **Get Books:** `GET /books`
- **Get Book by ID:** `GET /books/{id}`
- **Update Book:** `PUT /books/{id}`
- **Delete Book:** `DELETE /books/{id}`
- **Rent Book:** `POST /books/rent/{id}`
- **Return Book:** `POST /books/return/{id}`
- **Get Rentals:** `GET /rentals`

## Technologies Used

- Go (Golang)
- MongoDB
- [gofr.dev](https://gofr.dev) - Go framework for HTTP handling

## Contributing

Feel free to contribute to the project! Fork the repository, make your changes, and submit a pull request. Your contributions are highly appreciated.

