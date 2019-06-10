---
layout: post
title:  "우리의 번역을 Mattermost Translation Server에 반영하기"
categories: [ Contribution Guideline ]
image: "images/2019-06-10-young/preview.PNG"
author: young
---

![image](/2019-1-OSS-L4/images/2019-06-10-young/1.PNG)  
저희는 지금까지 ko.json 파일들의 내용을 번역하여 수정했습니다.  
또한 이 문장들은 [Mattermost Translation Server](https://translate.mattermost.com/ko/)에서도 전부 찾아볼 수 있습니다.  
실제 기여를 위해서는 이 Mattermost Translation Server에 반영해야 하므로 이에 관하여 포스트를 작성합니다.  

# 로컬에서의 번역 반영하기
저희가 수정한 파일은 3개입니다.  
mattermost-server/i18n/ko.json  
mattermost-webapp/i18n/ko.json  
mattermost-mobile/assets/base/i18n/ko.json  
이 중 mattermost-server/i18n/ko.json 파일은 json 구조가 다른 두 파일과 다르며 id를 명시하고 있고, 다른 두 파일은 "id":"translation" 형태로 되어 있습니다.  
Mattermost Translation Server에 우리의 번역을 반영하기 위해서는 이 id가 필요합니다.  
![image](/2019-1-OSS-L4/images/2019-06-10-young/2.PNG)  
먼저 번역을 반영할 문장이 있는 파일에서 id를 복사합니다.  
![image](/2019-1-OSS-L4/images/2019-06-10-young/3.PNG)  
이 id를 Mattermost Translation Server의 찾기 기능에서 사진과 같은 설정을 통해 검색합니다.  
![image](/2019-1-OSS-L4/images/2019-06-10-young/4.PNG)  
그러면 다음과 같이 저희가 찾는 문장이 나타나고 번역을 반영하여 제출 버튼을 누르시면 됩니다.  