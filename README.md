# **golang-todo-app: A Golang-Powered To-Do List Application**

## **Overview**
This golang-todo-app is a simple yet powerful to-do list application built with Golang. It showcases key backend development concepts and tools, including authentication with Auth0, structured logging, database management with PostgreSQL, caching with Redis, and email notifications using SendGrid.

---

## **Features**
- **User Authentication**: Secure login and registration via Auth0.
- **Task Management**: Create, view, update, and delete tasks.
- **Role-Based Access Control (RBAC)**: Manage access based on user roles.
- **Email Notifications**: Reminders and daily task summaries using SendGrid.
- **Caching**: Redis caching for frequently accessed data.
- **Tracing and Logging**:
  - Distributed tracing with OpenTelemetry.
  - Structured logging using Uber Zap.
- **Database Management**: PostgreSQL integration with migrations using Golang Migrate.

---

## **Tech Stack**

| **Component**        | **Technology**        |
|-----------------------|-----------------------|
| Language              | Golang               |
| Authentication        | Auth0                |
| Database              | PostgreSQL           |
| Caching               | Redis                |
| Logging               | Uber Zap             |
| Notifications         | SendGrid             |
| ORM                   | sqlc                 |
| Web Framework         | Go templates         |
| Tracing               | OpenTelemetry        |
| Containerization      | Docker               |

---

## **Installation and Setup**

### **Prerequisites**
- [Docker](https://www.docker.com/)
- [Go (1.19+)](https://golang.org/)
- Auth0 account
- SendGrid account

### **Steps to Run Locally**

1. **Clone the Repository**
   ```bash
   git clone https://github.com/henryhall897/golang-todo-app
   cd golang-todo-app

2. **Set Up Environment Variables Create a .env file in the root directory:**
    - AUTH0_DOMAIN=your-auth0-domain
    - AUTH0_CLIENT_ID=your-client-id
    - AUTH0_CLIENT_SECRET=your-client-secret
    - POSTGRES_URI=postgres://username:password@localhost:5432/golang-todo-app
    - REDIS_URI=redis://localhost:6379
    - SENDGRID_API_KEY=your-sendgrid-api-key
    - TRACING_URL=http://localhost:9411/api/v2/spans
    - LOG_LEVEL=debug
    - APP_PORT=8080

3. **Run Docker Containers Use docker-compose to start the app and its dependencies:**
    - docker-compose up --build

4. **Access the Application**
    - The application will be available at http://localhost:8080.

## **Usage**
### 1. Authentication:
- Register or log in via the web interface.
- Access token includes RBAC roles and permissions.
### 2. Task Management:

- Add new tasks with a title, description, and due date.
- Update or delete tasks as needed.
- Filter tasks by status (e.g., pending, completed).
### 3. Notifications:
- Enable email reminders for upcoming tasks.
- Receive daily summaries of pending tasks.

## **API Endpoints**

| **Endpoint**     | **Method** | **Description**              |
|-------------------|------------|------------------------------|
| `/auth/login`     | POST       | User login                  |
| `/auth/register`  | POST       | User registration           |
| `/tasks`          | GET        | Get all tasks               |
| `/tasks/:id`      | GET        | Get task by ID              |
| `/tasks`          | POST       | Create a new task           |
| `/tasks/:id`      | PUT        | Update an existing task     |
| `/tasks/:id`      | DELETE     | Delete a task               |
## **Contributing**

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch (`feature-name`).
3. Commit your changes.
4. Push the branch and create a pull request.

---

## **License**

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## **Contact**

For questions or feedback, feel free to reach out:

- **Email**: [henrydhall117@gmail.com](mailto:henrydhall117@gmail.com)
- **GitHub**: [henrydhall897](https://github.com/henrydhall897)
