中文 | [English](./README_EN.md)

# 推特媒體下載器

以用戶名為參數，下載用戶所有的推文的媒體文件（圖片、視頻）等

# 變更日志

- 2024-05-04

  - 支持視頻下載；
  - 保存下載記錄；
  - 支持增量式下載；
  - 支持從讀取文件中的下載鏈接進行下載

# 使用方法

- 配置

創建一個 setting.json 文件。填入 cookie。

```json
{
  "cookie": "auth_token=xxxx; ct0=xxxxx",
  "cookieComment": "上面cookie字段填入你的cookie，保證cookie有效，否則將無法正常使用"
}
```

- 如何獲取 cookie，示例

  - 網頁上登陸推特
  - F12 打開網頁調試器
  - 跟隨圖片步驟和提示複製 cookie

  ![cookie示例](https://img.outsider404.com/asuhe-blog-img/2024/05/46750ca579c1d92c37310ee9d07c932c.png)

- 運行

在 release 裏下載 main.exe 程序，點擊運行。
你自己用源碼編譯也可以。

- 使用用戶名進行下載

菜單中選擇第一項，并輸入用戶名。

![用戶名示例](https://img.outsider404.com/asuhe-blog-img/2024/05/ac2935726958b3416581cc34ab8e55af.png)

- 使用配置文件下載

文件名必須為`urls.json`。并且以數組形式放置需要下載的鏈接

```json
[
  "https://pbs.twimg.com/media/GMAXMobaYAAk3Ab.jpg",
  "https://pbs.twimg.com/media/GL_PbVhawAAwmxC.jpg"
]
```
