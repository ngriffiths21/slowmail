# HTTP Routes

### GET routes

- `/`, `/mail/inbox`: List and previews of received mail
- `/mail/inbox/{id}`: Read a single mail
- `/mail/draft/`: List and previews of drafts
- `/mail/draft/new`: Create a new mail (redirects to `draft/{id}`)
- `/mail/draft/{id}`: Edit a draft
- `/account/new`: Create a new account
- `/account/login`: Log in (with link to sign up). All routes redirect here if auth fails.
- `/account`: View and update account settings

### POST routes

- `/account/new`: Create new account
- `/account/login`: Log in
- `/mail/draft/{id}`: Save a mail
- `/mail/send/{id}`: Send a mail
