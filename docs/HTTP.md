# HTTP Routes

### GET routes

- `/`: Redirect to `/mail/folder/inbox`
- `/mail/folder/inbox`: List and previews of received mail
- `/mail/piece/{id}/read`: Read a single mail
- `/mail/folder/drafts/`: List and previews of drafts
- `/mail/piece/{id}/edit`: Edit a draft
- `/signup`: Create a new account
- `/login`: Log in (with link to sign up). All routes redirect here if auth fails.
- `/account`: View and update account settings

##### GET handlers

1. (`/mail/...`, `/account`) Check authentication
2. Load data
    - (`/mail/piece/...`) Single mail
    - (`/mail/folder/...`) List of mail
    - (`/account`) Account info
3. Render template and respond

### POST routes

- `/signup`: Create new account
- `/login`: Log in
- `/mail/piece/{id}`: Send a mail
- `/mail/new`: Create a new mail (creates and redirects to `/mail/piece/{id}/edit`)

##### POST handlers

1. (`/mail/...`, `/account`) Check authentication
2. Save data & retrieve any results
    - (`/signup`) Create account
    - (`/login`) Log in
    - (`/mail/...`) Create mail
    - (`/mail/piece/{id}`) Delete old draft
3. (`/signup`, `/login`) Set auth cookie
4. Redirect

### PUT routes

- `/mail/piece/{id}`: Save a draft
- `/account/`: Save account info

##### PUT handlers

1. Check authentication
2. Update data
    - `/mail/{id}/draft`: Update mail
    - `/account/`: Update account
3. Redirect
