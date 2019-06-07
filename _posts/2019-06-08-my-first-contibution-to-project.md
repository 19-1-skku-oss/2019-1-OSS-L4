---
layout: post
title:  "실제 Mattermost-Server에 Code로 기여하기 까지"
author: seunghyeon
categories: [ Code Contribution ]
image: assets/images/my-first-contribution/PR_Closed.png
---

## Mattermost Project에 내가 작성한 Code로 기여하는 과정

  해당 포스트는 오픈소스소프트웨어 실습에 참여하면서 팀 단위로 주제를 선정한 이후,실제로 issue를 통해 server분야 작업할 프로젝트를 선정하고 이에 저희가 수정한 Code를 commit하고 Pull Request를 통하여 저희가 작성한 Code가 Merge되어 Project에 기여햔 과정을 담았습니다.  

  Mattermost에 contribtue하는 과정은 아래 포스트를 통해 확인할 수 있습니다.
 [Contribution GuideLine] (https://19-1-skku-oss.github.io/2019-1-OSS-L4/Mattermost_Contribution-Guideline(1))

   Server 분야가 아닌 Mobile 분야의 PR을 통한 contribution은 아래 게시글을 통해 확인 가능합니다.
[Mobile PR] (https://19-1-skku-oss.github.io/2019-1-OSS-L4/mobile-pr/)

  저희가 프로젝트를 통해 해결한 이슈와 PR 링크는 다음과 같습니다.
`MM-15795`  [PR](https://github.com/mattermost/mattermost-server/pull/11000) /[Issue](https://github.com/mattermost/mattermost-server/issues/10937)
`MM-15799`  [PR](https://github.com/mattermost/mattermost-server/pull/11038) /[Issue](https://github.com/mattermost/mattermost-server/issues/10933)
`MM-15303`  [PR](https://github.com/mattermost/mattermost-server/pull/10927) /[Issue](https://github.com/mattermost/mattermost-server/issues/10714)
`MM-15304`  [PR](https://github.com/mattermost/mattermost-server/pull/10940) /[Issue](https://github.com/mattermost/mattermost-server/issues/10713)


## GO? Docker?

Mattermost 서버의 경우 `GO` 언어를 기반으로 동작하는데, 저희 모두 GO언어를 접해본 적이 없기 때문에 처음에는 모르는 언어로 진행되는 프로젝트에 Code로 기여할 수 있을까 하는 막막함과 두려움이 앞섰습니다. 하지만 Mattermost 프로젝트의 경우 이슈별로 난이도가 `Difficulty` tag를 통해 잘 나타나 있고, 같은 내용을 담은 이슈들과 Code를 확인할 수 있기 때문에 이를 참조한다면 저희도 할 수 있을 것이라 생각되어 조우림 팀장님의 도움을 받아 코드 기여에 도전해 보기로 하였습니다.

![image](/assets/images/my-first-contribution/issue_list.png)  


![image](/assets/images/my-first-contribution/PR_code.png)  

+ Migrate 이슈의 경우 함수의 구조를 파악하여 단순히 Migration하면 되기 때문에 전체적인 process가 동일합니다. 따라서 제가 맡았던 `MM-15957` 이슈를 토대로 과정을 기록하였습니다.

## Issue 할당 받기

 [Issue 게시글 사진]

  첫 단계는 이슈를 확인하고 아직 할당되지 않은, `Up For Grabs` Tag가 붙어 있는 이슈에 참여의사를 밝히는 일입니다. 각 Issue에는 다음과 같은 정보가 포함되어 있습니다.

  + Jira Ticket
  각 이슈는 MM-15795와 같이 할당된 번호가 있으며, 이와 같은 이름의 branch를 만들어 코드를 수정하고 Pull Request를 날린 후, 이를 관리자가 master branch에 merge하는 과정을 통해 프로젝트에 추가됩니다. 

  + Expected way to implement
  또한 issue에는 코드 수정이 어떤식으로 진행되어야 할지에 대한 대략적인 방법이 나와있기 때문에, 이를 그대로 따라서 수정하기만 해도 code 기여에 참여할 수 있습니다. 
  MM-15795의 경우 `store/sqlstore` 폴더 내의 `Session` store에서 `Getsessions` method를 수정하고, 수정 사항에 맞춰 Interface를 수정하라고 명시되어 있습니다.
  
  + Example
  Example에는 비슷한 내용을 담은 meger된 PR이 링크되어 있으며, 여기서 수정된 Code를 통해 어떤 식으로 바꿔야 하는지를 파악할 수 있습니다.

  위와 같은 내용을 확인한 후 issue에 자신이 이를 맡겠다는 comment를 남기는 것으로 issue를 할당 받을 수 있습니다.

## 본격적인 Code 수정

  Issue를 할당받아 코드 수정 후, 저희의 경우 Pull Request전에 저희 Team L4 repository에 수정된 코드를 미리 PR하여 서로 code를 확인해보는 시간을 가졌습니다. 이를 통해 정식으로 프로젝트에 Pull Request하기 전에 서로의 Code를 점검할 수 있었습니다.

  ![image](/assets/images/my-first-contribution/tp_pr.png)  


## Pull Request 

  Pull Request을 날리면 이에 대한 Code 리뷰와 함께 Feedback을 받을 수 있으며,리뷰 내용을 기반으로 수정 작업에 들어갑니다. 저 같은 경우 Code 수정 이후에도 Master branch와의 conflict 문제 등으로 인해 추가적인 수정을 해야 했습니다만, 결국 merge에 성공하여 생에 첫 기여에 성공하였습니다.

![image](/assets/images/my-first-contribution/pr_1.png)  
![image](/assets/images/my-first-contribution/pr_2.png)  
![image](/assets/images/my-first-contribution/pr_3.png)  

## 오픈소스 프로젝트를 하며 느낀 점

#### 코딩 스타일의 중요성

  저 같은 경우 개인적으로 코딩을 할 때는 코딩 스타일에 크게 신경을 쓰지는 않지만 오픈소스 프로젝트의 경우 코드의 가독성 및 통일성 또한 중요한 요소이기 때문에 줄 바꿈 하나, 띄어쓰기 하나에 신경을 써야했습니다. 이를 통해 개인 과제를 할 때에도 코딩 스타일에 신경쓰게 되는 계기가 되었습니다.

#### 오픈소스 프로젝트에 기여하는 뿌듯함

  비록 큰 수정은 아니었지만, 제가 작성한 코드가 큰 프로젝트의 일부로 merge 되었을 때 큰 성취감과 뿌듯함을 느낄 수 있었습니다. 이를 통해 왜 사람들이 오픈소스 프로젝트에 참여하는지를 알 수 있었습니다.

#### 협업의 중요성

  저 같은 경우 GO언어를 접해 본 적이 없어서 간단한 수정임에도 불구하고 어려움을 겪었는데, 조우림 팀장님을 비롯하여, mattermost를 담당하는 분들이 code에 대해 친절하고 정확한 피드백을 해주셔서 결국에는 code가 merge될 수 있었습니다. 서로 도와가며 code를 작성한다는 것 자체가 신선한 경험이었고, 저도 다른 사람들을 도와줄 수 있도록 열심히 해야겠다고 생각한 계기가 되었습니다.