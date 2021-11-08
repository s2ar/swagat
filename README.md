[![Build Status](https://scrutinizer-ci.com/g/s2ar/swagat/badges/build.png?b=main)](https://scrutinizer-ci.com/g/s2ar/swagat/build-status/main)
![Go Report](https://goreportcard.com/badge/github.com/s2ar/swagat)
![Repository Top Language](https://img.shields.io/github/languages/top/s2ar/swagat)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/s2ar/swagat/badges/quality-score.png?b=main)](https://scrutinizer-ci.com/g/s2ar/swagat/?branch=main)
![Lines of code](https://img.shields.io/tokei/lines/github/s2ar/swagat)
![Github Open Issues](https://img.shields.io/github/issues/s2ar/swagat)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/s2ar/swagat)
![Github Repository Size](https://img.shields.io/github/repo-size/s2ar/swagat)
![GitHub last commit](https://img.shields.io/github/last-commit/s2ar/swagat)
# Swagat service

## Описание
Websocket API Gateway. При запуске сервиса, он подписыватся на Bitmex уведомления, и транслирует клиентам согласно их подпискам.

http://localhost:9090/ открывает страницу клиента. На ней отображается котировки на которые подписан клиент, также кнопки управления подписками. Для подписки на все котировки, поле ввода нужно оставить пустым. Для подписки на несколько символов необходимо перечислить их через запятую 
![2021-11-08-b9c](https://user-images.githubusercontent.com/2817417/140735195-dde63526-b239-4b94-b8ba-ec462cc5e55d.png)

## Howto:
`make run`


