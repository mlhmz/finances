# Pockets

## Skills

### Skill Spec

Add the following skill

/spec feature-number

Brainstorm step by step; Ask questions if needed. Do the brainstorm in iterations, visualize in ASCII-Art if needed.

Get the features from /spec/roadmap.md

Write everything into a Brainstorm Markdown File, questions will be answered with [ ].

The final specification will be written into the spec.

Files:
/spec/{nr}_{summary}_brainstorm.md
/spec/{nr}_{summary}_spec.md

### Skill Dev

Add the following skill

/dev feature-number

Implement, write unit tests, write Cypress E2E tests, verify and fix until everything works.

Get the spec from /spec/{nr}_{summary}_spec.md

Write your progress /spec/{nr}_{summary}_progress.md

## Roadmap

Write a roadmap for my application. Don't decide anything; ask me if you have to. Step by step we will brainstorm the application and build specifications for the Application.

Divide the roadmap into features.

Write the roadmap into /spec/roadmap.md

### Features

- Authentication
    - Passwordless authentication via email
    - Use console as email confirmation for now
    - Use JWT for authentication, including refresh tokens
    - Users can use profile pictures, use initials for now
    - The login mask is always only the email
    - When email does not exist, show give option to type Full name, Currency (only show implemented currencies)
    - Password reset
- Multi currency with money pattern
    - Implement euro for now
- Multi-Tenancy
    - Own data for every user that uses the application
- Manual tracking of transactions
- Users have multiple account spaces
    - Can be named
    - Can be colored
- Users can track fix costs
    - Start of the fix cost
    - Schedule (once in a month on 16th or once in 3 months)
- Categorization of Transactions
    - New categories can be created (Lazy selection, create on selection if not existant)
    - Categories can be automatically suggested by an LLM with the Transaction title and description
- Dashboard and Reporting
    - Overview of total balance
    - Spending trends over time
- CSV export
- User audit log of changes in financial data
