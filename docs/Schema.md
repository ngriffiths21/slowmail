# Database Schema

### Tables

##### Table `mail`

- `mail_id` (integer primary key): Slow Mail id for mail
- `user_id` (integer not null): Slow Mail user ID
- `folder` (varchar(25)): Slow Mail folder
    - check folder in('inbox', 'drafts', 'sent', 'archive')
- `read` (tinyint not null): Boolean flag for read (set to 1 if anything other than unread inbox mail)
- `orig_date` (unsigned int not null): Date time received in mail header, in Unix seconds
- `date` (unsigned int not null): Date of delivery, rounded to midnight and given in Unix seconds
- `from_head` (text not null): Combined content of from, sender, and reply-to mail headers
    - check length(from_head) > 0
- `from_name` (varchar(40)): Display name of primary sender
- `from_addr` (varchar(255) not null): Email of primary sender
    - check length(from_addr) > 0
- `to_head` (text): Combined content of to and cc mail headers
- `message_id` (text unique not null): Message ID from mail header. If internal to slowmail, not necessary.
- `in_reply_to` (text): Content of in-reply-to header, with message IDs of parent message(s)
- `subject` (text): Content of subject header
- `content` (text): Mail body
- `multifrom` (tinyint not null): Boolean flag for more than one from address
- `multito` (tinyint not null): Boolean flag for more than one to address

##### Table `users`

- `user_id` (integer primary key): Slow Mail user ID
- `username` (varchar(40) unique not null): Email username
    - check length(username) > 0
- `password` (binary(64) not null): Hash of password. The hash of empty data is also considered invalid.
- `display_name` (varchar(40) not null): Display name
    - check length(display_name) > 0
- `recovery_addr` (varchar(255)): Recovery email (optional)

##### Table `sessions`

- `session_id` (varchar(11) unique not null): Session ID, 8 bytes in base64, which should be (securely) randomly generated
- `user_id` (integer not null): Slow Mail user ID
- `start_date` (unsigned int not null): Date time the session started, in UNIX seconds
- `ip` (varchar(40) not null): IP address of client
- `expiration` (unsigned int not null): Date time the session expires, in UNIX seconds

##### Table `drafts`

- `user_id` (integer not null): Slow Mail user ID of sender
- `recipient` (varchar(40) not null): Recipient address
- `subject` (text): Subject of message
- `content` (text): Content of message
- PRIMARY KEY (user_id, recipient)

### Data validation

The following data constraints are the responsibility of the client to enforce (implemented using HTML attributes, or when necessary, client-side JavaSript). If invalid data reaches the database driver, this is considered an application bug, not a user error.

- Data types (valid strings and numeric values)
- Minimums and maximums for string lengths and numeric values
- String patterns (e.g., a valid email, a strong password)

The following data constraints are the responsibility of the server to enforce, and when necessary, to handle gracefully by prompting the user for a different value:

- Unique fields (in particular, usernames)
