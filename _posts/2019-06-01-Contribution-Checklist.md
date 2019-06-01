---
layout: post
title:  "Contribution Checklist"
author: seunghyeon
categories: [ Code Contribution ]
image: assets/images/4.jpg
---

이 게시글은 Mattermost 에서 제공한 Contribution Checklist를 참고했습니다.

원문 : https://developers.mattermost.com/contribute/getting-started/contribution-checklist/


## Mattermost 이슈 확인을 통한 Contribution 시작하기, 그리고 Pull Request
https://19-1-skku-oss.github.io/2019-1-OSS-L4/mm-15303-15304/

이슈 확인 부터 PR까지의 전체적인 과정은 위 게시물을 참고하면 됩니다.

## Pull Request 전 다음과 같은 사항들을 체크해야합니다.

#### 1. CLA 등록하기
https://www.mattermost.org/mattermost-contributor-agreement/
CLA(Contributor License Agreement)에 등록하고, Mattermost의 Approved Contributor List에 이름이 올라가야 합니다.

#### 2. ticket 이 help wanted Github issue여야 합니다.
임의로 코드를 수정해서는 안되며, Mattermost의 github issue tab을 확인하고 자신이 담당할 이슈를 배정 받은 상태에서 작업에 들어간 코드여야 합니다.

#### 3. Code는 Mattermost Style Guide를 따라야 합니다.
https://docs.mattermost.com/developer/style-guide.html
이는 깃헙 repo 내에서 make check-sytle 명령어를 통해서도 확인 가능합니다.

#### 4. Code 수정 후 Mattermost에서 제공하는 test를 거쳐야 합니다.
https://developers.mattermost.com/contribute/webapp/end-to-end-tests/
- end-to-end tests for webapp

#### 5. 변경사항이 적용가능하다면, UI의 문자열은 localization files 에 포함되어 있어야 합니다.
../mattermost-server../en.json
../mattermost-webapp../en.json
../mattermost-mobile../en.json

#### 6. PR은 master branch를 대상으로 제출되어야 합니다.

#### 7. PR title은 JIRA 또는 Github Ticket ID로 시작해야 합니다.(MM-394 or GH - 394) 그리고 summary 부분을 채워 넣어야 합니다.

#### 8. PR이 한번 제출되면, 자동 빌드 과정이 PR을 통해 전달 됩니다.
- PR이후 자신의 branch에서 수정 및 commit을 거치면 PR이 자동으로 업데이트 된다는 의미입니다.
