---
layout: post
title:  "Mattermost 오픈소스 참여로 얻은 보상"
categories: [ MatterMost, Code Contribution, Pull Request ]
image: "images/2019-06-07-mattermug/mattermug.jpg"
author: woolim
---
# MatterMost 프로젝트 참여하기  
지금까지 저희 팀 페이지의 4가지 게시글을 통해 MatterMost 기여 방법에 대해 소개해드렸습니다.

[Mattermost 기여 가이드라인(1)](https://19-1-skku-oss.github.io/2019-1-OSS-L4/Mattermost_Contribution-Guideline(1)/)  
[Mattermost 번역 기여하기](https://19-1-skku-oss.github.io/2019-1-OSS-L4/Mattermost_Translation/)  
[Contribution Checklist](https://19-1-skku-oss.github.io/2019-1-OSS-L4/Contribution-Checklist/)  
[MatterMost 이슈 확인을 통한 Contribution 시작하기, 그리고 Pull Request](https://19-1-skku-oss.github.io/2019-1-OSS-L4/mm-15303-15304/)  

저는 코드 기여 위주로 활동하여 지금까지 총 7개 이슈 해결에 참여하였고,  
6개의 Pull Request를 날렸습니다.  
서버 3개의 경우 Pull Request가 전부 Merge 되었고, 모바일 3개의 경우 1개는 Merge 되었습니다.   
나머지 2건의 경우 Approve는 받았으나 어플리케이션 자체의 Library Migration으로 인해 Merge를 기다리는 중입니다.  
모바일 3개의 경우는 해당 PR의 내용에 대해 다음 게시물에서 자세히 설명해 볼 예정입니다.  

오픈소스 프로젝트에 제대로 참여해본건 이번이 처음이라 많은 것을 느꼈는데,  
이번엔 제가 MatterMost 프로젝트에 기여하면서 얻은 것에 대해 설명해 보려고 합니다.  

# 새로운 개발 트렌드, 언어, 플랫폼 접하기  
![image](/2019-1-OSS-L4/images/2019-06-07-mattermug/docker.jpg)  
MatterMost의 경우 서버는 docker를 사용하고 backend 언어는 go를 사용하고 있습니다.  
docker와 go 모두 접해보지 않았기 때문에 저의 개발 pool을 넓힐 수 있는 좋은 기회를 가지게 되었습니다.
서버 PR의 경우 함수를 Migration 하는 3가지를 해결했는데, 단순히 해당 코드만 보는 것이 아니라
MatterMost 프로젝트 구조나 빌드 과정 등에 대해 자세히 알 수 있었습니다.  

# 코드 리뷰  
![image](/2019-1-OSS-L4/images/2019-06-07-mattermug/code_review.png)  
무료로 코드 리뷰를 받을 수 있다는 것도 하나의 장점입니다.  
예를 들어 제가 처음에 해결한 이슈 [MM-13879](https://github.com/mattermost/mattermost-mobile/pull/2832)의 경우 구현에는 성공하였지만,  
이전에 ReactNative에 대한 경험도 없고 테스트 주도 개발에도 전혀 관심이 없었기 때문에 테스트 코드를 전혀 작성하지 않았었습니다.  
그러나, 리뷰를 통해서 코드에서 부족한 점과 고칠 부분을 명확히 알 수 있었고, 이슈를 쉽게 해결할 수 있었습니다.  

# MatterMug 
![image](/2019-1-OSS-L4/images/2019-06-07-mattermug/mattermug.jpg)  
이 부분은, MatterMost 프로젝트의 특징이라고 할 수도 있는데,
Code Contribution에 참여하거나 여러가지 조건 중 하나를 만족하면 아래와 같이 MatterMost 머그컵 "MatterMug"를 줍니다.  

![image](/2019-1-OSS-L4/images/2019-06-07-mattermug/mattermug_email.png)  
코드 기여를 어느정도 하다보니 저에게도 위와 같이 "MatterMug"를 위한 정보를 보내달라는 메일이 도착했습니다.  
사실 별 것 아닐 수도 있지만, MatterMost 프로젝트에 대한 기여를 기념하는 것과, 기여에 대한 보답을 실제로 확인할 수 있다는 것이
큰 의미가 있는 것 같습니다.