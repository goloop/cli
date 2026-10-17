[![Go Reference](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/goloop/cli) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](https://github.com/goloop/cli/blob/v1/LICENSE) [![Stay with Ukraine](https://img.shields.io/static/v1?label=Stay%20with&message=Ukraine%20♥&color=ffD700&labelColor=0057B8&style=flat)](https://u24.gov.ua/)

# cli

`goloop` - невеликий CLI-інструмент для скафолдингу й інспекції goloop-first
Go-проєктів. Це звичайний диспетчер поверх `os.Args` без сторонніх залежностей,
який ніколи не комітить, не пушить і не деплоїть.

## Встановлення

```bash
go install github.com/goloop/cli@latest
```

Бінарник називається `goloop`.

## Команди

```text
goloop new <name> [--module path] [--dir path] [--force]
goloop check [dir]     скан на сліди ШІ-асистента (ненульовий код при знахідці)
goloop doctor          перевірка наявності потрібних інструментів
goloop status          підсумок модуля в поточній теці
goloop version         версія CLI
```

### new

Створює мінімальний goloop-first сервіс (`main.go` на базі
[`goloop/app`](https://github.com/goloop/app), плюс `go.mod`, `.env`,
`.gitignore` і `README.md`). Ім'я - малі літери, цифри й одиночні дефіси.
Відмовляється перезаписувати наявну теку без `--force`.

### check

Обходить теку й повідомляє про будь-які сліди ШІ-асистента (пропускаючи `.git`,
`node_modules`, `vendor`, `dist`). Використовуйте в CI, щоб тримати
product-репозиторії чистими. Ненульовий код при знахідці.

### doctor

Повідомляє, які з очікуваних інструментів (`go`, `git`, `docker`, `node`)
встановлено.

## Межі

MVP покриває `new`, `check`, `doctor`, `status` і `version`. Багатша
оркестрація (мульти-репо топології, deploy-скафолдинг, compose smoke) свідомо
поза першою версією.

## Документація

- Англійський довідник: [DOC.md](DOC.md)
- Український довідник: [DOC.UK.md](DOC.UK.md)

## Ліцензія

MIT - див. [LICENSE](LICENSE).

## Реліз v1.0.0

check повертає помилку обходу відсутнього або недоступного каталогу. Новий сервіс використовує app v1.0.0. Реліз v1 зберігає поточний import path і публічний API. Майбутні несумісні зміни потребуватимуть нової головної версії.
