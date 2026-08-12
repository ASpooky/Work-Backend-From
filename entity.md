# エンティティ定義

前提: 個人利用のみ（認証なし、NAS/ローカルホスティング）。
`user_id` は将来の複数ユーザー拡張に備えて **User / Workspace にのみ**持たせ、
Goal / DailyTask は Workspace 経由で所有者を辿る（全テーブルに撒かない）。

## User

| field       | type      | note                                   |
|-------------|-----------|-----------------------------------------|
| id          | uuid      | 単一ユーザー運用では固定値でも可         |
| name        | string    | 表示名                                   |
| created_at  | datetime  |                                          |

## Workspace

| field       | type      | note                                   |
|-------------|-----------|-----------------------------------------|
| id          | uuid      |                                          |
| user_id     | uuid (FK) | User.id                                 |
| name        | string    | 例: "private", "work"                   |
| created_at  | datetime  |                                          |

## Goal

| field                | type              | note                                             |
|----------------------|-------------------|---------------------------------------------------|
| id                    | uuid              |                                                    |
| workspace_id          | uuid (FK)         | Workspace.id                                      |
| title                 | string            |                                                    |
| detail                | text              |                                                    |
| achievement_condition | text              | 達成条件                                          |
| end_date              | date              | 目標終了日                                        |
| mode                  | enum(strict/want) | strict: 未達で end_date がずれる / want: ずれない |
| status                | enum(active/achieved/abandoned) |                                     |
| created_at            | datetime          |                                                    |

## DailyTask

| field       | type      | note                                             |
|-------------|-----------|-----------------------------------------------------|
| id          | uuid      |                                                      |
| goal_id     | uuid (FK) | Goal.id                                             |
| date        | date      | 対象日                                              |
| content     | text      | その日にやること                                    |
| done        | boolean   | 達成フラグ                                          |
| created_at  | datetime  |                                                      |

## 確定事項

- strict モードで未達のとき、`end_date` は1日ずれる（未達1日 = +1日）
- Goal の `status` は全 DailyTask 達成で自動的に achieved になる（アラート選択式にはしない）

## 実装しながら決める（レコード形は上記のまま固定、運用だけ後決め）

- DailyTask の生成タイミング（Goal作成時に一括 / 都度 / 日付の箱だけ先行生成で中身は都度、の3案で C 案が有力候補）
