import { useState } from 'react'
import { applyTheme, loadTheme, type ThemePreference } from '../theme'

const API_KEY_STORAGE_KEY = 'goal-tracker:ai-api-key'

function SettingsPage() {
  const [theme, setTheme] = useState<ThemePreference>(() => loadTheme())
  const [apiKey, setApiKey] = useState(() => localStorage.getItem(API_KEY_STORAGE_KEY) ?? '')
  const [saved, setSaved] = useState(false)

  function handleThemeChange(next: ThemePreference) {
    setTheme(next)
    applyTheme(next)
  }

  function handleSaveApiKey() {
    localStorage.setItem(API_KEY_STORAGE_KEY, apiKey)
    setSaved(true)
    setTimeout(() => setSaved(false), 1500)
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
          AI連携 <span className="badge-unimplemented">未実装</span>
        </h2>
        <p className="settings-note">
          ここで保存した API キーは今のところブラウザのローカルストレージに保存されるだけで、まだどの機能からも参照されていません。
        </p>
        <div className="settings-api-key">
          <input
            type="password"
            placeholder="sk-..."
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
          />
          <button type="button" onClick={handleSaveApiKey}>
            保存
          </button>
          {saved && <span className="settings-saved">保存しました</span>}
        </div>
      </section>
    </>
  )
}

export default SettingsPage
