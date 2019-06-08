---
layout: post
title:  "Mattermost 기여 가이드라인(2)"
categories: [ Contribution Guideline ]
author: chaehyeon
image: "https://user-images.githubusercontent.com/49435910/59152744-b3014880-8a85-11e9-9b3c-b0350e0e94f5.png"
---
이 게시글은 Mattermost 에서 제공한 Contributions Without Ticket을 참고했습니다.

원문: [https://developers.mattermost.com/contribute/getting-started/contributions-without-ticket/](https://developers.mattermost.com/contribute/getting-started/contributions-without-ticket/)

# Contributions Without Ticket

원하는 ticket이 없더라도 사소한 수정이나 개선으로 기여를 할 수 있습니다.
기존의 방식에 대해 변화가 작은 것이라면, Help Wanted ticket이 없어도 오류 혹은 점진적 개선에 대한 20줄 이하의 코드수정 PR 제출이 받아들여 질 수 있습니다.
ticket이 없는 pull requests는 먼저 core team product manager에게 검토됩니다.

### 사소한 수정이나 개선에 대한 예
[Fix a formatting error in help text](https://github.com/mattermost/mattermost-server/pull/5640)
![20190609_062052](https://user-images.githubusercontent.com/49435910/59152436-ce1c8a00-8a7e-11e9-9e9a-a71aabf3896a.png)


[Fix success typo in Makefile](https://github.com/mattermost/mattermost-server/pull/5809)
![20190609_062226](https://user-images.githubusercontent.com/49435910/59152441-fb693800-8a7e-11e9-9dfe-32765223127f.png)


[Fix broken Cancel button in Edit Webhooks screen](https://github.com/mattermost/mattermost-server/pull/5612)
![20190609_062427](https://user-images.githubusercontent.com/49435910/59152453-42572d80-8a7f-11e9-9085-db40b00367c5.png)


[Fix Android app crashing when saving user notification settings](https://github.com/mattermost/mattermost-mobile/pull/364)
![20190609_062513](https://user-images.githubusercontent.com/49435910/59152462-5dc23880-8a7f-11e9-88e2-1a6b4fc06df0.png)


[Fix recent mentions search not working](https://github.com/mattermost/mattermost-server/pull/5878)
![20190609_062557](https://user-images.githubusercontent.com/49435910/59152471-75012600-8a7f-11e9-8ba9-cf45f45ba748.png)


RP을 검토하는 Core committer가 판단했을 때 방식이나 사용자 기대에 큰 변화를 미치는 것이라고 생각하면 PR을 거부할 자격이 있습니다. 또한 그것은 core team에 의해 Help Wanted ticket으로 열리도록 요구됩니다.
Core team과 함께 Help Wanted ticket을 열기 위한 방법은 [feature idea 포럼에서 회의하기](https://www.mattermost.org/feature-ideas/) 혹은 [GitHub repository에서 issue 올리기](https://github.com/mattermost/mattermost-server/issues/new)가 있습니다.
또한 Contributors나 Developers channel에서도 대화해 볼 수 있습니다.

### Core committers
Core committer는 Mattermost 저장소에 merge할 프로젝트 관리자입니다. 이들은 pull requests를 검토하고, Mattermost 개발자 커뮤니티를 구축하며, Mattermost의 기술 비전을 제시합니다.
만약 질문이 있거나 도움이 필요하다면 core committer에게 요청해보세요.

[Mattermost의 core committers list](https://developers.mattermost.com/contribute/getting-started/core-committers/)
