# `go-booru-bot`

### Description

This is a straightforward Telegram bot for sending images from **derpibooru.org** directly within your chats. It uses **inline queries**, so you can search for and share images without leaving the conversation.

---

### Status

#### Working Features :heavy_check_mark:

* **Inline Queries**: Search for images by typing `@go-booru-bot [your query]`.
    * **Paging**: Get more results by adding `#[page number]` to your query (e.g., `@go-booru-bot twilight sparkle #2`). Each page displays **25 results**.
* **/featured Command**: Use this command to quickly view the most popular or "featured" images from derpibooru.org.

#### Planned Improvements :construction:

* **Animated Content**: Add support for handling and sending **GIFs** and **MP4s**.
* **Dynamic Loading**: Implement a better way to load more results directly in the inline query window, so you don't have to manually specify page numbers.
