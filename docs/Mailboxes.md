# Mailbox documentation

How mail flows through the application, and how a user interacts with it.

### General concepts

Mail are organized in *conversations* between two users, where each user only has the mail that the other has sent to them. As a user, it is possible to view an inbox with the day's mail from each conversation, a folder of archived conversations, or a folder of drafts.

### Drafts

When composed mail gets saved, it goes in drafts.

- Only a single draft can exist per recipient.

### Inbox and archive

When mail is sent, it is routed from one user's drafts to the other user's inbox.

- For each conversation (unique sender), only the most recent mail will be displayed in the inbox or archive.
- In the inbox and archive, mail can have a status of read or unread.

Inbox:

- In the inbox, only the current day's mail is displayed.
- At a certain time each day (I will set 2pm to start), mail with that date becomes displayable to the user. Before that time the previous day is displayed.

Archive:

- The archive contains conversations that either are older than a day, or were manually archived by the user.
- All conversations not in the inbox are displayed regardless of how old.

### Conversations

- Mail is read in a page containing the whole conversation.
- If there is a draft response, it is displayed at the top. Otherwise, the user can begin a new response.
