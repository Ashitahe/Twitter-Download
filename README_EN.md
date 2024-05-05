[中文](README.md) | English

# Twitter Media Downloader

Download all media files (images, videos, etc.) from a user's tweets using the username as a parameter.

# Changelog

- 2024-05-05

  - Generate CSV record file

- 2024-05-04

  - Support for video downloads;
  - Save download records;
  - Support for incremental downloads;
  - Support for downloading from links in a read file.

- How to Use

  - Configuration

Create a `setting.json` file. Fill in the cookie.

```json
{
  "cookie": "auth_token=xxxx; ct0=xxxxx",
  "cookieComment": "Fill in your cookie in the cookie field above, ensuring the cookie is valid, otherwise, it will not function properly."
}
```

- How to obtain a cookie, example:

  - Log in to Twitter on a webpage.

  - Open the web debugger with F12.

  - Follow the steps and hints in the picture to copy the cookie.

![cookie example](https://img.outsider404.com/asuhe-blog-img/2024/05/46750ca579c1d92c37310ee9d07c932c.png)

- Running

Download the `main.exe` program from the release, and click to run.
You can also compile from the source code yourself.

- Download using a username

Select the first item in the menu and enter the username.

![username example](https://img.outsider404.com/asuhe-blog-img/2024/05/ac2935726958b3416581cc34ab8e55af.png)

- Download using a configuration file

The file name must be `urls.json`. Place the links to be downloaded in an array format.

```json
[
  "https://pbs.twimg.com/media/GMAXMobaYAAk3Ab.jpg",
  "https://pbs.twimg.com/media/GL_PbVhawAAwmxC.jpg"
]
```
