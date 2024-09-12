# HTTP Routes

### GET routes

- `/`: Redirect to `/mail/folder/inbox/`
- `/mail/folder/inbox/`: List and previews of unopened mail
- `/mail/folder/archive/`: List and previews of all received mail
- `/mail/conv/{id}/read/`: Read a conversation
- `/mail/folder/drafts/`: List and previews of drafts
- `/mail/compose/`: Compose page
- `/mail/draft/{id}/edit/`: Work on a draft (same as compose page)
- `/signup/`: Create a new account
- `/login/`: Log in (with link to sign up). All routes redirect here if auth fails.
- `/account/`: View and update account settings

##### GET handlers

1. (`/mail/...`, `/account/`) Check authentication
2. Load data
    - (`/mail/piece/...`) Single mail
    - (`/mail/folder/...`) List of mail
    - (`/account/`) Account info
3. Render template and respond

### POST routes

- `/signup/`: Create new account
- `/login/`: Log in
- `/mail/conv/{id}/send/`: Send a reply
- `/mail/compose/`: Save a newly composed draft
- `/mail/compose/send/`: Send a new mail
- `/mail/conv/{id}/save/`: Save a draft reply
- `/account/`: Save account info

##### POST handlers

1. (`/mail/...`, `/account/`) Check authentication
2. Save data & retrieve any results
    - (`/signup/`) Create account
    - (`/login/`) Log in
    - (`/mail/compose/`) Save mail as draft
    - (`/mail/compose/send/`) Create and send mail
    - (`/mail/conv/{id}/save/`): Create or update draft reply
    - (`/mail/conv/{id}/send/`): Delete old draft and send reply
    - (`/account/`): Update account
3. (`/signup`, `/login`) Set auth cookie
4. Redirect
