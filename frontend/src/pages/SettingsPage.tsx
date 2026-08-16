import { useEffect, useState } from 'react'
import { api } from '../api'
import { applyTheme, loadTheme, type ThemePreference } from '../theme'

function SettingsPage() {
  const [theme, setTheme] = useState<ThemePreference>(() => loadTheme())
  const [aiEnabled, setAiEnabled] = useState<boolean | null>(null)

  useEffect(() => {
    api
      .aiStatus()
      .then((s) => setAiEnabled(s.enabled))
      .catch(() => setAiEnabled(false))
  }, [])

  function handleThemeChange(next: ThemePreference) {
    setTheme(next)
    applyTheme(next)
  }

  return (
    <>
      <section>
        <h2>デザイン</h2>
        <div className="settings-theme-options">
          {(['system', 'light', 'dark'] as const).map((option) => (
            <label key={option} className="settings-radio">
              <input
                type="radio"
                name="theme"
                checked={theme === option}
                onChange={() => handleThemeChange(option)}
              />
              {option === 'system' ? 'システムに合わせる' : option === 'light' ? 'ライト' : 'ダーク'}
            </label>
          ))}
        </div>
      </section>

      <section>
        <h2>
          AI連携{' '}
          {aiEnabled === true && <span className="badge-enabled">有効</span>}
          {aiEnabled === false && <span className="badge-unimplemented">無効</span>}
        </h2>
        {aiEnabled === false && (
          <p className="settings-note">
            バックエンドに <code>GEMINI_API_KEY</code> 環境変数が設定されていないため無効です。サーバーを起動する前に
            設定してください（<code>GEMINI_MODEL</code> でモデルも変更できます、未設定時は gemini-3.7-flash）。
            APIキーはブラウザ側には一切保存されません。
          </p>
        )}
        {aiEnabled === true && (
          <p className="settings-note">
            サイドバーの「AIと計画」から、目標の壁打きとタスクへの落とし込みができます。
          </p>
        )}
      </section>
    </>
  )
}

export default SettingsPage
