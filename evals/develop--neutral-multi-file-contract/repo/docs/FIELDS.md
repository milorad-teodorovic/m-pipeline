# Field catalogue

| Field | Go name | JSON key | Required | Constraint |
|---|---|---|---|---|
| Customer ID | `ID` | `id` | yes | Starts with `cus_`, then 1 or more lowercase letters or digits. |
| Email | `Email` | `email` | yes | Contains exactly one `@`, with text on both sides. |
| Display name | `Name` | `name` | yes | 1 to 80 characters. |
| Phone number | `Phone` | `phone` | no | E.164: `+`, then 8 to 15 digits, and the first digit is not `0`. No spaces or other characters. |
| Marketing opt-in | `MarketingOptIn` | `marketing_opt_in` | no | Boolean. |
