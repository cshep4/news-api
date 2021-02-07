# News API

News feed service written in Go which can retrieve news articles from various providers for specified categories.

The REST API for the news retriever is described below.

## Get Feed

### Request

`GET /`

    curl --location --request GET 'localhost:8080?provider=bbc&limit=56&offset=3'

### Response

    {
        "provider": "bbc",
        "items": [
            {
                "category": "uk",
                "provider": "bbc",
                "title": "Covid vaccinations: Wales leads the UK on first vaccine dose rate",
                "link": "https://www.bbc.co.uk/news/uk-wales-55855220",
                "description": "More than 550,000 people in the top priority groups have been given first doses of Covid vaccines.",
                "thumbnail": "https://news.bbcimg.co.uk/nol/shared/img/bbc_news_120x60.gif",
                "pubDate": "2021-02-06T20:47:21Z"
            }
        ],
        "limit": 1
    }

## Get Feed by Category

### Request

`GET /{category}`

    curl --location --request GET 'localhost:8080/uk?provider=bbc&limit=56&offset=3'

### Response

    {
        "category": "uk",
        "provider": "bbc",
        "items": [
            {
                "category": "uk",
                "provider": "bbc",
                "title": "Covid vaccinations: Wales leads the UK on first vaccine dose rate",
                "link": "https://www.bbc.co.uk/news/uk-wales-55855220",
                "description": "More than 550,000 people in the top priority groups have been given first doses of Covid vaccines.",
                "thumbnail": "https://news.bbcimg.co.uk/nol/shared/img/bbc_news_120x60.gif",
                "pubDate": "2021-02-06T20:47:21Z"
            }
        ],
        "limit": 1
    }