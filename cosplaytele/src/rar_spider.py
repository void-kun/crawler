import scrapy
import json


class RarSpider(scrapy.Spider):
    name = "rar_spider"
    start_urls = ["https://cosplaytele.com/"]

    def media_parse(self, res):
        try:
            download_link = res.xpath(
                '//div[@id="download_link"]//a[@id="downloadButton"]'
            ).attrib["href"]
            self.add_downloaded_list(download_link)
        except Exception as e:
            print(e)

    def item_parse(self, res):
        try:
            media_url = res.xpath(
                "//div[@class='entry-content single-page']//p//a[@rel='nofollow noopener']"
            ).attrib["href"]
            yield scrapy.Request(media_url, callback=self.media_parse)
        except Exception as e:
            print(e)

    def parse(self, res):
        for item in res.xpath("//div[@id='post-list']//div[@class='col post-item']"):
            next_page = item.css(".plain").attrib["href"]
            yield scrapy.Request(next_page, callback=self.item_parse)

        next_page = res.xpath("//a[@class='next page-number']").attrib["href"]
        yield scrapy.Request(next_page, callback=self.parse)

    def add_downloaded_list(self, url: str) -> bool:
        try:
            with open("urls.txt", "a") as f_out:
                f_out.write(url + "\n")

            return True
        except Exception as e:
            print("\n\nerror : ", e)
            return False
