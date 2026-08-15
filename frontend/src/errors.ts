export function toUserMessage(err: unknown): string {
  console.error(err)
  return '通信に失敗しました。しばらくしてからもう一度お試しください。'
}
