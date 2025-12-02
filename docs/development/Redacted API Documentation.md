# Redacted API Documentation

The JSON API provides an easily parse-able interface to Redacted. Below is the list of information available, the arguments that can be passed to it, and the format of the results.

Questions about the API can be answered in `#red-dev`.  
Bugs are best directed to the [Bug Reports subforum](/forums.php?action=viewforum&forumid=13).

**Using the API bestows upon you a certain level of trust and responsibility. Abusing or using this API for malicious purposes is a bannable offense and will not be taken lightly.**

**Standard Login:**
- Refrain from making more than five (5) requests every ten (10) seconds.

**API Key Login:**
- Refrain from making more than ten (10) requests every ten (10) seconds.

Both of these rate limits apply on a per-user basis, not IP. This means if you have multiple scripts authenticating to the API with both methods, once you are over the `5r/10s` limit the Standard Login script will receive HTTP 429 responses while the API Key Login script will continue to function. Upon exceeding the rate limit for your method of authentication, you will receive a HTTP 429 error and a `Retry-After` header. The format of the header is `Retry-After: x` where x is the integer representation of the number of seconds remaining in your current API burst window.

## Authentication

To make requests to the Redacted API, you must be authenticated. Users can authenticate in two ways:

- Fetching a cookie through the login form at `https://redacted.sh/login.php`.
- Generating an API key in their user settings and sending a `Authorization: %apikey%` header with each request.
- Certain endpoints will be documented as being API key only. This is to prevent session riding / CSRF. In this situation you **must** authenticate with an API key.

For users with 2FA enabled, it may be easier to utilize the API key authentication option.

## Scope

When generating API keys, the scope of its access should be restricted to only grant access to the required areas of the site needed by the tool or application. This means specifying permissions that limit the API key's access to only the essential endpoints necessary for the tool to function. This ensures that, even if your API key is compromised, its limited scope will prevent unauthorized access to other parts of the site.

Scope access is defined below:

**API Scope permissions:**

### No Scope Required
- Index
- User Stats
- Torrent Search
- Logchecker
- Similar Artists
- Announcements

### User Scope
- User Profile
- Messages (Inbox, Conversation, Send PM)
- User Search
- Bookmarks
- Subscriptions
- User Torrents (Seeding, Leeching, Uploaded, Snatched)
- Notifications

### Torrent Scope
- Top 10
- Artist
- Upload Torrent
- Download Torrent
- Torrent Group (Group Details, Group Edit, Add Tag)
- Torrent (Torrent Details, Torrent Log Files, Torrent Edit)
- Collages (Collage Details, Adding Release to Collage)

### Request Scope
- Requests (Request Search, Request Fill, Request Details)

### Forum Scope
- Forum (Category View, Forum View, Thread View)

### Wiki Scope
- Wiki

## Changelog

Most recent updates:
- API Scopes defined

## Mappings

There are instances where the API returns an integer lookup. Those mappings will be included below.

**ReleaseType:**
```
1: "Album"
3: "Soundtrack"
5: "EP"
6: "Anthology"
7: "Compilation"
9: "Single"
11: "Live album"
13: "Remix"
14: "Bootleg"
15: "Interview"
16: "Mixtape"
17: "Demo"
18: "Concert Recording"
19: "DJ Mix"
21: "Unknown"
1021: "Produced By"
1022: "Composition"
1023: "Remixed By"
1024: "Guest Appearance"
```

Certain parameters such as `filter_cat[]`, `release[]`, `bitrates[]`, `formats[]`, and `media[]` expect mapping IDs, not strings. The relevant mappings are as below:

**filter_cat[]:**
```
0: "Music"
1: "Applications"
2: "E-Books"
3: "Audiobooks"
4: "E-Learning Videos"
5: "Comedy"
6: "Comics"
```

**release[]:**
```
1: "Album"
3: "Soundtrack"
5: "EP"
6: "Anthology"
7: "Compilation"
9: "Single"
11: "Live album"
13: "Remix"
14: "Bootleg"
15: "Interview"
16: "Mixtape"
17: "Demo"
18: "Concert Recording"
19: "DJ Mix"
21: "Unknown"
```

**bitrates[]:**
```
0: "192"
1: "APS (VBR)"
2: "V2 (VBR)"
3: "V1 (VBR)"
4: "256"
5: "APX (VBR)"
6: "V0 (VBR)"
7: "320"
8: "Lossless"
9: "24bit Lossless"
10: "Other"
```

**formats[]:**
```
0: "MP3"
1: "FLAC"
2: "AAC"
3: "AC3"
4: "DTS"
```

**media[]:**
```
0: "CD"
1: "DVD"
2: "Vinyl"
3: "Soundboard"
4: "SACD"
5: "DAT"
6: "Cassette"
7: "WEB"
8: "Blu-Ray"
```

## Outline

All request URLs are in the form: `ajax.php?action=<ACTION>`

All the JSON returned is in the form:
```json
{
  "status" : "success",
  "response" : {
    // Response data.
  }
}
```

If the request is invalid, or a problem occurs, the `status` will be `failure`. In this case the value of `response` is `undefined`.

## Index

**URL:** `ajax.php?action=index`

**Arguments:** None

**Response format:**
Note the addition of the `api_version` reply. This can be parsed by scripts needing to ensure the availability of the new/updated endpoints. This version number may be incremented as future changes are made.

```json
{
    "status": "success",
    "response": {
        "username": "dr4g0n",
        "id": 469,
        "authkey": "redacted",
        "passkey": "redacted",
        "api_version": "redacted-v2.0",
        "notifications": {
            "messages": 0,
            "notifications": 9000,
            "newAnnouncement": false,
            "newBlog": false
        },
        "userstats": {
            "uploaded": 585564424629,
            "downloaded": 177461229738,
            "ratio": 3.29,
            "requiredratio": 0.6,
            "class": "VIP"
        }
    }
}
```

## User Profile

**URL:** `ajax.php?action=user`

**Arguments:**
- `id` - id of the user to display

**Response format:**
```json
{
    "status": "success",
    "response": {
        "username": "dr4g0n",
        "avatar": "http://v0lu.me/rubadubdub.png",
        "isFriend": false,
        "profileText": "",
        "bbProfileText": "",
        "profileAlbum": {
            "id": "53",
            "name": "A Charlie Brown Christmas",
            "review": ""
        },
        "stats": {
            "joinedDate": "2007-10-28 14:26:12",
            "lastAccess": "2012-08-09 00:17:52",
            "uploaded": 585564424629,
            "downloaded": 177461229738,
            "ratio": 3.3,
            "requiredRatio": 0.6
        },
        "ranks": {
            "uploaded": 98,
            "downloaded": 95,
            "uploads": 85,
            "requests": 0,
            "bounty": 79,
            "posts": 98,
            "artists": 0,
            "overall": 85
        },
        "personal": {
            "class": "VIP",
            "paranoia": 0,
            "paranoiaText": "Off",
            "donor": true,
            "warned": false,
            "enabled": true,
            "passkey": "redacted"
        },
        "community": {
            "posts": 863,
            "torrentComments": 13,
            "collagesStarted": 0,
            "collagesContrib": 0,
            "requestsFilled": 0,
            "requestsVoted": 13,
            "perfectFlacs": 2,
            "uploaded": 29,
            "groups": 14,
            "seeding": 309,
            "leeching": 0,
            "snatched": 678,
            "invited": 7
        }
    }
}
```

## User Stats

**URL:** `ajax.php?action=community_stats`

**Arguments:**
- `userid` - id of the user to display

**Response format:**
```json
{
    "status": "success",
    "response": {
        "leeching": 0,
        "seeding": "1,413",
        "snatched": "1,574",
        "usnatched": "1,556",
        "downloaded": "1,993",
        "udownloaded": "1,729",
        "seedingperc": 91,
        "seedingsize": "418.36 GB"
    }
}
```

## Messages

### Inbox

**URL:** `ajax.php?action=inbox`

**Arguments:**
- `page` - page number to display (default: 1)
- `type` - one of: inbox or sentbox (default: inbox)
- `sort` - if set to *unread* then unread messages come first
- `search` - filter messages by search string
- `searchtype` - one of: subject, message, user

**Response format:**
```json
{
    "status": "success",
    "response": {
        "currentPage": 1,
        "pages": 3,
        "messages": [
            {
                "convId": 3421929,
                "subject": "1 of your torrents has been deleted for inactivity",
                "unread": false,
                "sticky": false,
                "forwardedId": 0,
                "forwardedName": "",
                "senderId": 0,
                "username": "",
                "donor": false,
                "warned": false,
                "enabled": true,
                "date": "2012-06-12 00:54:01"
            }
        ]
    }
}
```

### Conversation

**URL:** `ajax.php?action=inbox&type=viewconv`

**Arguments:**
- `id` - id of the message to display

**Response format:**
```json
{
    "status": "success",
    "response": {
        "convId": 3421929,
        "subject": "1 of your torrents has been deleted for inactivity",
        "sticky": false,
        "messages": [
            {
                "messageId": 4507261,
                "senderId": 0,
                "senderName": "System",
                "sentDate": "2012-06-12 00:54:01",
                "bbBody": "One of your uploads has been deleted for being unseeded...",
                "body": "One of your uploads has been deleted for being unseeded..."
            }
        ]
    }
}
```

### Send PM

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=send_pm`

**POST Arguments:**
- `toid` - user to send the PM to
- `convid` - conversation to send the pm under (optional)
- `subject` - subject of the conversation (required if convid is not specified)
- `body` - body of the message

**Response format:**
```json
{
    "status": "success",
    "response": "A conversation has been created with id: 321"
}
```

## Top 10

**URL:** `ajax.php?action=top10`

**Arguments:**
- `type` - one of: torrents, tags, users (default: torrents)
- `limit` - one of 10, 100, 250 (default: 10)

**Response format:**
```json
{
    "status": "success",
    "response": [
        {
            "caption": "Most Active Torrents Uploaded in the Past Day",
            "tag": "day",
            "limit": 10,
            "results": [
                {
                    "torrentId": 30194226,
                    "groupId": 72268716,
                    "artist": "2 Chainz",
                    "groupName": "Based on a T.R.U. Story",
                    "groupCategory": 0,
                    "groupYear": 2012,
                    "remasterTitle": "Deluxe Edition",
                    "format": "MP3",
                    "encoding": "V0 (VBR)",
                    "hasLog": false,
                    "hasCue": false,
                    "media": "CD",
                    "scene": true,
                    "year": 2012,
                    "tags": ["hip.hop"],
                    "snatched": 135,
                    "seeders": 127,
                    "leechers": 5,
                    "data": 17242225550
                }
            ]
        }
    ]
}
```

## User Search

**URL:** `ajax.php?action=usersearch`

**Arguments:**
- `search` - The search term.
- `page` - page to display (default: 1)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "currentPage": 1,
        "pages": 1,
        "results": [
            {
                "userId": 469,
                "username": "dr4g0n",
                "donor": true,
                "warned": false,
                "enabled": true,
                "class": "VIP"
            }
        ]
    }
}
```

## Bookmarks

**URL:** `ajax.php?action=bookmarks&type=<Type>`

**Arguments:**
- `type` - one of torrents, artists (default: torrents)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "bookmarks": [
            {
                "id": 71843824,
                "name": "Spacejams",
                "year": 2010,
                "recordLabel": "Hospital Records",
                "catalogueNumber": "NHS178CD",
                "tagList": "drum_and_bass electronic",
                "releaseType": "1",
                "vanityHouse": false,
                "image": "http://whatimg.com/i/09930203236341542660.jpg",
                "torrents": [
                    {
                        "id": 29043412,
                        "groupId": 71843824,
                        "media": "CD",
                        "format": "FLAC",
                        "encoding": "Lossless",
                        "remasterYear": 0,
                        "remastered": false,
                        "remasterTitle": "",
                        "remasterRecordLabel": "",
                        "remasterCatalogueNumber": "",
                        "scene": false,
                        "hasLog": true,
                        "hasCue": true,
                        "logScore": 100,
                        "fileCount": 15,
                        "freeTorrent": false,
                        "size": 563078107,
                        "leechers": 0,
                        "seeders": 26,
                        "snatched": 142,
                        "time": "2010-11-13 21:25:10",
                        "hasFile": 29043412
                    }
                ]
            }
        ]
    }
}
```

## Subscriptions

**URL:** `ajax.php?action=subscriptions`

**Arguments:**
- `showunread` - 1 to show only unread, 0 for all subscriptions (default: 1)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "threads": [
            {
                "forumId": 20,
                "forumName": "Technology",
                "threadId": 218,
                "threadTitle": "Post Your Desktop",
                "postId": 3844686,
                "lastPostId": 4149355,
                "locked": false,
                "new": true
            }
        ]
    }
}
```

## Forums

### Category View

**URL:** `ajax.php?action=forum&type=main`

**Response format:**
```json
{
    "status": "success",
    "response": {
        "categories": [
            {
                "categoryID": 1,
                "categoryName": "Site",
                "forums": [
                    {
                        "forumId": 19,
                        "forumName": "Announcements",
                        "forumDescription": "If you don't like the news, go out and make some of your own.",
                        "numTopics": 338,
                        "numPosts": 84368,
                        "lastPostId": 4148491,
                        "lastAuthorId": 331548,
                        "lastPostAuthorName": "Isocline",
                        "lastTopicId": 150195,
                        "lastTime": "2012-08-08 15:03:18",
                        "specificRules": [],
                        "lastTopic": "Whataroo 2012!",
                        "read": false,
                        "locked": false,
                        "sticky": false
                    }
                ]
            }
        ]
    }
}
```

### Forum View

**URL:** `ajax.php?action=forum&type=viewforum&forumid=<Forum Id>`

**Arguments:**
- `forumid` - id of the forum to display
- `page` - the page to display (default: 1)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "forumName": "Announcements",
        "specificRules": [],
        "currentPage": 1,
        "pages": 7,
        "threads": [
            {
                "topicId": 150195,
                "title": "Whataroo 2012!",
                "authorId": 168713,
                "authorName": "Steve096",
                "locked": false,
                "sticky": false,
                "postCount": 552,
                "lastID": 4148491,
                "lastTime": "2012-08-08 15:03:18",
                "lastAuthorId": 331548,
                "lastAuthorName": "Isocline",
                "lastReadPage": 0,
                "lastReadPostId": 0,
                "read": false
            }
        ]
    }
}
```

### Thread View

**URL:** `ajax.php?action=forum&type=viewthread&threadid=<Thread Id>&postid=<Post Id>`

**Arguments:**
- `threadid` - id of the thread to display
- `postid` - response will be the page including the post with this id
- `page` - page to display (default: 1)
- `updatelastread` - set to 1 to not update the last read id (default: 0)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "forumId": 7,
        "forumName": "The Lounge",
        "threadId": 159925,
        "threadTitle": "Women with short hair",
        "subscribed": false,
        "locked": false,
        "sticky": false,
        "currentPage": 1,
        "pages": 1,
        "posts": [
            {
                "postId": 4146433,
                "sticky": false,
                "addedTime": "2012-08-07 18:38:19",
                "bbBody": "Are so much sexier...",
                "body": "Are so much sexier...",
                "editedUserId": 0,
                "editedTime": "",
                "editedUsername": "",
                "author": {
                    "authorId": 310550,
                    "authorName": "Z0M813",
                    "paranoia": ["collages+", "collagecontribs+"],
                    "artist": false,
                    "donor": false,
                    "warned": false,
                    "avatar": "http://whatimg.com/i/vmrol8.jpeg",
                    "enabled": true,
                    "userTitle": ""
                }
            }
        ]
    }
}
```

## Artist

**URL:** `ajax.php?action=artist&id=<Artist Id>`

**Arguments:**
- `id` - artist's id
- `artistname` - Artist's Name
- `artistreleases` - if set, only include groups where the artist is the main artist.

**Response format:**
```json
{
    "status": "success",
    "response": {
        "id": 1460,
        "name": "Logistics",
        "notificationsEnabled": false,
        "hasBookmarked": true,
        "image": "http://img120.imageshack.us/img120/3206/logiop1.jpg",
        "body": "",
        "vanityHouse": false,
        "tags": [
            {
                "name": "breaks",
                "count": 3
            }
        ],
        "similarArtists": [],
        "statistics": {
            "numGroups": 125,
            "numTorrents": 443,
            "numSeeders": 3047,
            "numLeechers": 95,
            "numSnatches": 28033
        },
        "torrentgroup": [
            {
                "groupId": 72189681,
                "groupName": "Fear Not",
                "groupYear": 2012,
                "groupRecordLabel": "Hospital Records",
                "groupCatalogueNumber": "NHS209CD",
                "tags": ["breaks", "drum.and.bass", "electronic", "dubstep"],
                "releaseType": 1,
                "groupVanityHouse": false,
                "hasBookmarked": false,
                "torrent": [
                    {
                        "id": 29991962,
                        "groupId": 72189681,
                        "media": "CD",
                        "format": "FLAC",
                        "encoding": "Lossless",
                        "remasterYear": 0,
                        "remastered": false,
                        "remasterTitle": "",
                        "remasterRecordLabel": "",
                        "remasterCatalogueNumber": "",
                        "scene": true,
                        "hasLog": false,
                        "hasCue": false,
                        "logScore": 0,
                        "fileCount": 19,
                        "freeTorrent": false,
                        "size": 527749302,
                        "leechers": 0,
                        "seeders": 20,
                        "snatched": 55,
                        "time": "2012-04-14 15:57:00",
                        "hasFile": 29991962
                    }
                ]
            }
        ],
        "requests": [
            {
                "requestId": 172667,
                "categoryId": 1,
                "title": "We Are One (Nu:logic Remix)/timelapse",
                "year": 2012,
                "timeAdded": "2012-02-07 03:44:39",
                "votes": 3,
                "bounty": 217055232
            }
        ]
    }
}
```

## Torrents

### Torrents Search

**URL:** `ajax.php?action=browse&searchstr=<Search Term>`

**Arguments:**
- `searchstr` - string to search for
- `page` - page to display (default: 1)
- `taglist`, `tags_type`, `order_by`, `order_way`, `filter_cat`, `freetorrent`, `vanityhouse`, `scene`, `haslog`, `releasetype`, `media`, `format`, `encoding`, `artistname`, `filelist`, `groupname`, `recordlabel`, `cataloguenumber`, `year`, `remastertitle`, `remasteryear`, `remasterrecordlabel`, `remastercataloguenumber` - as in advanced search

**Response format:**
```json
{
    "status": "success",
    "response": {
        "currentPage": 1,
        "pages": 3,
        "results": [
            {
                "groupId": 410618,
                "groupName": "Jungle Music / Toytown",
                "artist": "Logistics",
                "tags": ["drum.and.bass", "electronic"],
                "bookmarked": false,
                "vanityHouse": false,
                "groupYear": 2009,
                "releaseType": "Single",
                "groupTime": 1339117820,
                "maxSize": 237970,
                "totalSnatched": 318,
                "totalSeeders": 14,
                "totalLeechers": 0,
                "torrents": [
                    {
                        "torrentId": 959473,
                        "editionId": 1,
                        "artists": [
                            {
                                "id": 1460,
                                "name": "Logistics",
                                "aliasid": 1460
                            }
                        ],
                        "remastered": false,
                        "remasterYear": 0,
                        "remasterCatalogueNumber": "",
                        "remasterTitle": "",
                        "media": "Vinyl",
                        "encoding": "24bit Lossless",
                        "format": "FLAC",
                        "hasLog": false,
                        "logScore": 79,
                        "hasCue": false,
                        "scene": false,
                        "vanityHouse": false,
                        "fileCount": 3,
                        "time": "2009-06-06 19:04:22",
                        "size": 243680994,
                        "snatches": 10,
                        "seeders": 3,
                        "leechers": 0,
                        "isFreeleech": false,
                        "isNeutralLeech": false,
                        "isFreeload": false,
                        "isPersonalFreeleech": false,
                        "trumpable": false,
                        "canUseToken": true
                    }
                ]
            }
        ]
    }
}
```

### Upload Torrent

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=upload`

This endpoint expects a **POST** method

**Arguments**
- `dryrun` - (bool) Only return the derived information from the posted data without actually uploading the torrent.
- `file_input` - (file) .torrent file contents
- `type` - (int) index of category (Music, Audiobook, ...)
- `artists[]` - (str)
- `importance[]` - (int) index of artist type (Main, Guest, Composer, ...) **One-indexed!**
- `title` - (str) Album title
- `year` - (int) Album "Initial Year"
- `releasetype` - (int) index of release type (Album, Soundtrack, EP, ...)
- `unknown` - (bool) Unknown Release
- `remaster_year` - (int) Edition year
- `remaster_title` - (str) Edition title
- `remaster_record_label` - (str) Edition record label
- `remaster_catalogue_number` - (str) Edition catalog number
- `scene` - (bool) is this a scene release?
- `format` - (str) MP3, FLAC, etc
- `bitrate` - (str) 192, Lossless, Other, etc
- `other_bitrate` - (str) bitrate if Other
- `vbr` - (bool) other_bitrate is VBR
- `logfiles[]` - (files) ripping log files
- `extra_file_#`
- `extra_format[]`
- `extra_bitrate[]`
- `extra_release_desc[]`
- `vanity_house` - (bool) is this a Vanity House release?
- `media` - (str) CD, DVD, Vinyl, etc
- `tags` - (str)
- `image` - (str) link to album art
- `album_desc` - (str) Album description **(ignored if new torrent is merged or added to existing group)**
- `release_desc` - (str) Release (torrent) description
- `desc` - (str) Description for non-music torrents
- `groupid` - (int) torrent groupID (ie album) this belongs to
- `requestid` - (int) requestID being filled

**Response format:**
```json
{
    "status": "success",
    "response": {
        "private": "...",
        "source": "...",
        "requestid": "...",
        "torrentid": "...",
        "groupid": "...",
        "newgroup": "..."
    }
}
```

### Download Torrent

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=download`

**Arguments**
- `id` - TorrentID to download.
- `usetoken` (optional) - Default: `0`. Set to `1` to spend a FL token. Will fail if a token cannot be spent on this torrent for any reason.

**Response format:**
On success, response is a .torrent file with `content-type: application/x-bittorrent;`. Failures return the usual JSON formatted error.

### Torrent Group

#### Group Details

**URL:** `ajax.php?action=torrentgroup&id=<Torrent Group Id>`

**Arguments:**
- `id` - torrent's group id
- `hash` - hash of a torrent in the torrent group (must be uppercase)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "group": {
            "wikiBody": "",
            "bbBody": "",
            "wikiImage": "https://ptpimg.me/r4eata.jpg",
            "id": 72189681,
            "name": "Fear Not",
            "year": 2012,
            "recordLabel": "Hospital Records",
            "catalogueNumber": "NHS209CD",
            "releaseType": 1,
            "categoryId": 1,
            "categoryName": "Music",
            "time": "2012-05-02 07:39:30",
            "collages": [],
            "personalCollages": [],
            "vanityHouse": false,
            "musicInfo": {
                "composers": [],
                "dj": [],
                "artists": [
                    {
                        "id": 1460,
                        "name": "Logistics"
                    }
                ],
                "with": [
                    {
                        "id": 25351,
                        "name": "Alice Smith"
                    }
                ],
                "conductor": [],
                "remixedBy": [],
                "producer": []
            }
        },
        "torrents": [
            {
                "id": 29991962,
                "media": "CD",
                "format": "FLAC",
                "encoding": "Lossless",
                "remastered": false,
                "remasterYear": 0,
                "remasterTitle": "",
                "remasterRecordLabel": "",
                "remasterCatalogueNumber": "",
                "scene": true,
                "hasLog": false,
                "hasCue": false,
                "logScore": 0,
                "fileCount": 19,
                "size": 527749302,
                "seeders": 20,
                "leechers": 0,
                "snatched": 55,
                "hasSnatched": true,
                "trumpable": true,
                "lossyWebApproved": false,
                "lossyMasterApproved": true,
                "freeTorrent": false,
                "isNeutralleech": false,
                "isFreeload": false,
                "time": "2012-04-14 15:57:00",
                "description": "...",
                "fileList": "...",
                "filePath": "Logistics-Fear_Not-CD-FLAC-2012-TaBoo",
                "userId": 567,
                "username": null
            }
        ]
    }
}
```

#### Group Edit

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=groupedit&id=<Torrent Id>`

**GET Arguments:**
- `id` - group id

**POST Arguments:** **At least one must be included**
- `summary` - Summary of edit changes

**Optional:** **At least one must be included**
- `body` - (str) Album description
- `image` - (str) Link to album art
- `releasetype` - (int) Index of release type (Album, Soundtrack, EP, ...)
- `groupeditnotes` - (str) Any editing notes that should be displayed when a user is making an edit.

**Response format:**
```json
{
    "status": "success",
    "response": "Torrent Group 1 was edited by Starlord (Body: 'testing1231231as' -> 'testing1231231asd')"
}
```

#### Add Tag

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=addtag`

This endpoint expects a **POST** method

**Arguments**
- `groupid` - Torrent GroupID to add the tag to.
- `tagname` - Tags to be added. Format: tagname1,tagname2

**Response format:**
```json
{
    "status": "success",
    "response": {
        "added": [],
        "voted": ["testing", "testing2"],
        "rejected": []
    }
}
```

### Torrent

#### Torrent Details

**URL:** `ajax.php?action=torrent&id=<Torrent Id>`

**Arguments:**
- `id` - torrent's id
- `hash` - torrent's hash (must be uppercase)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "group": {
            "wikiBody": "",
            "wikiImage": "http://whatimg.com/i/ralpc.jpg",
            "id": 72189681,
            "name": "Fear Not",
            "year": 2012,
            "recordLabel": "Hospital Records",
            "catalogueNumber": "NHS209CD",
            "releaseType": 1,
            "categoryId": 1,
            "categoryName": "Music",
            "time": "2012-05-02 07:39:30",
            "vanityHouse": false,
            "musicInfo": {
                "composers": [],
                "dj": [],
                "artists": [
                    {
                        "id": 1460,
                        "name": "Logistics"
                    }
                ],
                "with": [
                    {
                        "id": 25351,
                        "name": "Alice Smith"
                    }
                ],
                "conductor": [],
                "remixedBy": [],
                "producer": []
            }
        },
        "torrent": {
            "id": 29991962,
            "infoHash": "C88921B1F46363E05BC151BF836D2F2A0E87F8DF",
            "media": "CD",
            "format": "FLAC",
            "encoding": "Lossless",
            "remastered": false,
            "remasterYear": 0,
            "remasterTitle": "",
            "remasterRecordLabel": "",
            "remasterCatalogueNumber": "",
            "scene": true,
            "hasLog": false,
            "hasCue": false,
            "logScore": 0,
            "ripLogIds": [],
            "fileCount": 19,
            "size": 527749302,
            "seeders": 20,
            "leechers": 0,
            "snatched": 55,
            "freeTorrent": false,
            "isNeutralleech": false,
            "isFreeload": false,
            "time": "2012-04-14 15:57:00",
            "description": "...",
            "fileList": "...",
            "filePath": "Logistics-Fear_Not-CD-FLAC-2012-TaBoo",
            "userId": 567,
            "username": null
        }
    }
}
```

**Note:** The `group` object in the Torrent Details response uses `id`, `name`, and `year` fields (not `groupId`, `groupName`, `groupYear`).

#### Torrent Log Files

**URL:** `ajax.php?action=riplog&id=<Torrent Id>&logid=<Log Id>`

**Arguments:**
- `id` - torrent's id
- `logid` - logfile id

**Response format:**
```json
{
    "status": "success",
    "response": {
        "id": 4470175,
        "log": "...",
        "log_sha256": "46cefcccbd7e9146b2a144276ae68640570904531c0b1a3ec66e43f86ed01fe4",
        "logid": 990753,
        "score": 100,
        "checksum": true
    }
}
```

#### Torrent Edit

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=torrentedit&id=<Torrent Id>`

**GET Arguments:**
- `id` - torrent's id

**(Optional) POST Arguments:** **At least one must be included**
- `format` - (str) MP3, FLAC, etc
- `media` - (str) CD, DVD, Vinyl, etc
- `bitrate` - (str) 192, Lossless, Other, etc
- `release_desc` - (str) Release (torrent) description
- `remaster_year` - (int) Edition year
- `remaster_title` - (str) Edition title
- `remaster_record_label` - (str) Edition record label
- `remaster_catalogue_number` - (str) Edition catalog number
- `scene` - (bool) Is this a scene release?
- `unknown` - (bool) Unknown Release

**Response format:**
```json
{
    "status": "success",
    "response": "Torrent 1 (testing) in group 1 was edited by Starlord (Description: 'testing1231231as' -> 'testing1231231asd'"
}
```

## Logchecker

**URL:** `ajax.php?action=logchecker`

**Arguments:**
None

**One of the next two arguments must be provided. In the case both are provided the uploaded log will take precedence**

**Post Parameters:**
- `pastelog` - The log file you would like checked.

**File Upload:**
- `log` - The log file you would like checked.

**Response format:**
```json
{
    "status": "success",
    "response": {
        "score": 59,
        "issues": [
            "\"Defeat audio cache\" should be yes (-10 points)",
            "Gap handling must be appended to previous track (-10 points)",
            "Null samples should be used in CRC calculations (-1 point)",
            "Test and copy was not used (-20 points)"
        ]
    }
}
```

## User Torrents (Seeding, Leeching, Uploaded, Snatched)

**URL:** `ajax.php?action=user_torrents&id=<User ID>&type=<Torrent Type>&limit=<Results Limit>&offset=<Torrents Offset>`

**Arguments:**
- `id` - request id
- `type` - type of torrents to display options are: `seeding` `leeching` `uploaded` `snatched`
- `limit` - number of results to display (default: 500)
- `offset` - number of results to offset by (default: 0)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "seeding": [
            {
                "groupId": "4",
                "name": "If You Have Ghost",
                "torrentId": "4",
                "artistName": "Ghost B.C.",
                "artistId": "4"
            }
        ]
    }
}
```

## Requests

### Request Search

**URL:** `ajax.php?action=requests&search=<term>&page=<page>&tags=<tags>`

**Arguments:**
- `search` - search term
- `page` - page to display (default: 1)
- `tags` - tags to search by (comma separated)
- `tags_type` - `0` for any, `1` for match all
- `show_filled` - Include filled requests in results - `true` or `false` (default: false).
- `filter_cat[]`, `releases[]`, `bitrates[]`, `formats[]`, `media[]` - as used on requests.php and as defined in Mappings

If no arguments are specified then the most recent requests are shown.

**Response format:**
```json
{
    "status": "success",
    "response": {
        "currentPage": 1,
        "pages": 1,
        "results": [
            {
                "requestId": 185971,
                "requestorId": 498,
                "requestorName": "Satan",
                "timeAdded": "2012-05-06 15:43:17",
                "lastVote": "2012-06-10 20:36:46",
                "voteCount": 3,
                "bounty": 245366784,
                "categoryId": 1,
                "categoryName": "Music",
                "artists": [
                    [
                        {
                            "id": "1460",
                            "name": "Logistics"
                        }
                    ]
                ],
                "tags": {
                    "551": "japanese",
                    "1630": "video.game"
                },
                "title": "Fear Not",
                "year": 2012,
                "image": "http://whatimg.com/i/ralpc.jpg",
                "description": "Thank you kindly.",
                "catalogueNumber": "",
                "releaseType": "",
                "bitrateList": "1",
                "formatList": "Lossless",
                "mediaList": "FLAC",
                "logCue": "CD",
                "isFilled": false,
                "fillerId": 0,
                "fillerName": "",
                "torrentId": 0,
                "timeFilled": ""
            }
        ]
    }
}
```

### Request Fill

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=requestfill`

This endpoint expects a **POST** method

**Arguments**
- `requestid` - RequestID to fill.
- `link` - Permalink to torrent which fills the request.
- `torrentid` - TorrentID which fills the request.

Either TorrentID or Link are required.

**Response format:**
```json
{
    "status": "success",
    "response": {
        "requestid": 119451,
        "requestname": "Margari's Kid - We Are Ghosts Now",
        "fillerid": 31580,
        "fillername": "Saturn",
        "torrentid": 2327967,
        "bounty": 11832131584
    }
}
```

### Request Details

**URL:** `ajax.php?action=request&id=<Request Id>`

**Arguments:**
- `id` - request id
- `page` - page of the comments to display (default: last page)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "requestId": 80983,
        "requestorId": 75670,
        "requestorName": "brontosaurus",
        "timeAdded": "2010-01-08 03:12:39",
        "canEdit": false,
        "canVote": true,
        "minimumVote": 20971520,
        "voteCount": 765,
        "lastVote": "2012-08-08 20:37:24",
        "topContributors": [
            {
                "userId": 75670,
                "userName": "brontosaurus",
                "bounty": 1254160859136
            }
        ],
        "totalBounty": 1489901312461,
        "categoryId": 1,
        "categoryName": "Music",
        "title": "4th Studio Album",
        "year": 2012,
        "image": "",
        "description": "This request is for a proper rip to FLAC at 24 bits / 96 kHz...",
        "musicInfo": {
            "composers": [],
            "dj": [],
            "artists": [
                {
                    "id": 431,
                    "name": "Daft Punk"
                }
            ],
            "with": [],
            "conductor": [],
            "remixedBy": [],
            "producer": []
        },
        "catalogueNumber": "",
        "releaseType": 0,
        "releaseName": "Unknown",
        "bitrateList": "1",
        "formatList": "24bit Lossless",
        "mediaList": "FLAC",
        "logCue": "Vinyl",
        "isFilled": false,
        "fillerId": 0,
        "fillerName": "0",
        "torrentId": 0,
        "timeFilled": "0",
        "tags": ["electronic", "house", "french"],
        "comments": [
            {
                "postId": 63934,
                "authorId": 209372,
                "name": "verysofttoiletpaper",
                "donor": true,
                "warned": false,
                "enabled": true,
                "class": "Member",
                "addedTime": "2012-07-10 09:02:34",
                "avatar": "http://majastevanovich.files.wordphpss.com/2009/10/a20toilet20paper.jpg",
                "comment": "Can someone explain what is the attractiveness of a vinyl rip...",
                "editedUserId": 0,
                "editedUsername": "",
                "editedTime": ""
            }
        ],
        "commentPage": 18,
        "commentPages": 18
    }
}
```

## Collages

### Collage Details

**URL:** `ajax.php?action=collage&id=<Collage Id>`

**Arguments:**
- `id` - collage's id

**Optional arguments:**
- `showonlygroups` - if set, does not return torrent information.

**Response format:**
```json
{
    "status": "success",
    "response": {
        "id": 32,
        "name": "Days Of Being Wild",
        "description": "French producers and djs...",
        "creatorID": 138,
        "deleted": false,
        "collageCategoryID": 4,
        "collageCategoryName": "Label",
        "locked": false,
        "maxGroups": 0,
        "maxGroupsPerUser": 0,
        "hasBookmarked": false,
        "subscriberCount": 0,
        "torrentGroupIDList": ["2788", "3077", "3143"],
        "torrentgroups": [
            {
                "id": "2788",
                "name": "Conducting The Band",
                "year": "2012",
                "categoryId": "1",
                "recordLabel": "Days Of Being Wild",
                "catalogueNumber": "010",
                "vanityHouse": "0",
                "tagList": "electronic",
                "releaseType": "5",
                "wikiImage": "",
                "musicInfo": {
                    "composers": [],
                    "dj": [],
                    "artists": [
                        {
                            "id": 4656,
                            "name": "Catalepsia"
                        }
                    ],
                    "with": [],
                    "conductor": [],
                    "remixedBy": [
                        {
                            "id": 4907,
                            "name": "Club Bizarre"
                        }
                    ],
                    "producer": []
                },
                "torrents": [
                    {
                        "torrentid": 4186,
                        "media": "WEB",
                        "format": "MP3",
                        "encoding": "320",
                        "remastered": true,
                        "remasterYear": 2012,
                        "remasterTitle": "",
                        "remasterRecordLabel": "Days Of Being Wild",
                        "remasterCatalogueNumber": "010",
                        "scene": false,
                        "hasLog": false,
                        "hasCue": false,
                        "logScore": 0,
                        "fileCount": 5,
                        "size": 81945530,
                        "seeders": 2,
                        "leechers": 0,
                        "snatched": 11,
                        "freeTorrent": false,
                        "isNeutralleech": false,
                        "isFreeload": false,
                        "time": "2016-11-25 18:54:04"
                    }
                ]
            }
        ]
    }
}
```

### Adding Release to Collage

**This endpoint is restricted to API Key Authentication**

**URL:** `ajax.php?action=addtocollage&collageid=<Collage Id>`

**GET Arguments:**
- `collageid` - Collage ID

**POST Arguments:**
- `groupids` - (str) List of groupids to add. **Format:** 1,2,3

**Response format:**
```json
{
    "status": "success",
    "response": {
        "collage": "1",
        "groupsadded": ["1", "2"],
        "groupsrejected": [],
        "groupsduplicated": []
    }
}
```

## Notifications

**URL:** `ajax.php?action=notifications&page=<Page>`

**Arguments:**
- `page` - page number to display (default: 1)

**Response format:**
```json
{
    "status": "success",
    "response": {
        "currentPages": 1,
        "pages": 105,
        "numNew": 0,
        "results": [
            {
                "torrentId": 30194383,
                "groupId": 71944561,
                "groupName": "You Are a Tourist",
                "groupCategoryId": 1,
                "torrentTags": "alternative indie",
                "size": 12279586,
                "fileCount": 2,
                "format": "MP3",
                "encoding": "320",
                "media": "WEB",
                "scene": false,
                "groupYear": 2011,
                "remasterYear": 0,
                "remasterTitle": "",
                "snatched": 2,
                "seeders": 3,
                "leechers": 0,
                "notificationTime": "2012-08-08 21:24:15",
                "hasLog": false,
                "hasCue": false,
                "logScore": 0,
                "freeTorrent": false,
                "isNeutralleech": false,
                "isFreeload": false,
                "logInDb": false,
                "unread": false
            }
        ]
    }
}
```

## Similar Artists

**URL:** `ajax.php?action=similar_artists&id=<Artist ID>&limit=<Limit>`

**Arguments**
- `id` - id of artist
- `limit` - maximum number of results to return (fewer might be returned)

**Response format:**
```json
[
    {
        "id": 8307,
        "name": "Fairmont",
        "score": 200
    },
    {
        "id": 3693,
        "name": "Paul Kalkbrenner",
        "score": 200
    }
]
```

## Announcements

**URL:** `ajax.php?action=announcements&page=<Page>&perpage=<Results Per Page>`

**(Optional) GET Arguments**
- `page` - (int) What page should be returned. **Default:** 1
- `perpage` - (int) How many news posts & blog posts are requested per page. **Default:** 5
- `order_way` - (str) asc or desc. **Default:** desc
- `order_by` - (str) id, title, body, time. **Default:** time

**Response format:**
```json
{
    "status": "success",
    "response": {
        "announcements": [
            {
                "newsId": 73,
                "title": "Up With the Birds (Spring Update)",
                "bbBody": "...",
                "body": "...",
                "newsTime": "2020-04-10 14:55:47"
            }
        ]
    }
}
```

## Wiki

**URL:** `ajax.php?action=wiki`

**Arguments**
- `id` - id of wiki article
- `name` - alias of wiki article

**Response format:**
```json
{
    "status": "success",
    "response": {
        "title": "Changing Username",
        "bbBody": "**Can I rename my account?**\r\nPlease do not ask us...",
        "body": "<strong>Can I rename my account?</strong><br />\r\nPlease do not ask us...",
        "language": "English",
        "aliases": "changingusername,renameacc,username",
        "languages": [
            {
                "name": "English",
                "id": "76"
            }
        ],
        "authorID": 2,
        "authorName": "Nala",
        "date": "2017-04-20 22:19:48",
        "revision": 1
    }
}
```

## Documentation Pending

These endpoints have been implemented, but the docs have not been fully written yet. Watch this space!

**There may be significant bugs in the documentation and/or implementation of these endpoints. If you use anything in this section you do so at your own risk.**

## Unofficial projects that utilize the API

Many projects which utilize the API can be found in the [Sandbox Forum](/forums.php?action=viewforum&forumid=8). Examples include:

- Python - [[Python] Sahel: Rate limit respecting, cache-optional basic Python API for Redacted](/forums.php?action=viewthread&threadid=56402)
- Go - https://github.com/autobrr/autobrr
- Go - https://github.com/s0up4200/redactedhook

For reference purposes **only**, developers may wish to consult these older projects based on the vanilla what.cd API. We've tried hard to make our enhancements reverse compatible, but some scripts have better parsers than others. Since these scripts are too old to support our API key system, they should not be used as-is for general development.

- Python - https://github.com/isaaczafuta/whatapi
- Java - https://github.com/Gwindow/WhatAPI
- Ruby - https://github.com/chasemgray/RubyGazelle
- Javascript - https://github.com/deoxxa/whatcd
- C# - https://github.com/frankston/WhatAPI
- PHP - https://github.com/Jleagle/php-gazelle
- Go - https://github.com/kdvh/whatapi

