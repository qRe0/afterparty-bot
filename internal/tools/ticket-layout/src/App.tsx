import { useEffect, useMemo, useRef, useState } from 'react'
import type { PointerEvent } from 'react'
import ticketUrl from '../../../../assets/ticket.png'
import fontUrl from '../../../../assets/font.ttf?url'
import {
  CENTER_TOLERANCE,
  DEFAULT_LAYOUT,
  HANDLE_HIT_RADIUS,
  HANDLE_SIZE,
  MAX_FONT_SIZE,
  buildEnvSnippet,
  clamp,
  formatOffset,
  ggBaselineOffset,
  measureTextBounds,
  normalizeHexColor,
  visualCenterOffset,
} from './layout'
import type { LayoutState } from './layout'

type DragState =
  | {
      mode: 'move'
      pointerId: number
      startX: number
      startY: number
      startPosX: number
      startPosY: number
    }
  | {
      mode: 'resize'
      pointerId: number
      startDistance: number
      startFontSize: number
    }

type CopyState = 'idle' | 'copied' | 'failed'

const GUIDE_VISIBLE_MS = 1800

function App() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const dialogRef = useRef<HTMLDialogElement>(null)
  const dragRef = useRef<DragState | null>(null)
  const guideTimerRef = useRef<number | null>(null)
  const [measureCanvas] = useState(() => document.createElement('canvas'))

  const [image, setImage] = useState<HTMLImageElement | null>(null)
  const [imageSize, setImageSize] = useState({ width: 1110, height: 638 })
  const [fontReady, setFontReady] = useState(false)
  const [assetError, setAssetError] = useState<string | null>(null)
  const [layout, setLayout] = useState<LayoutState>(DEFAULT_LAYOUT)
  const [sample, setSample] = useState('1234')
  const [colorDraft, setColorDraft] = useState(DEFAULT_LAYOUT.color)
  const [colorError, setColorError] = useState<string | null>(null)
  const [copyState, setCopyState] = useState<CopyState>('idle')
  const [guidesVisible, setGuidesVisible] = useState(false)

  const textBounds = useMemo(() => {
    if (!image || !fontReady || !sample) return null

    const ctx = measureCanvas.getContext('2d')
    if (!ctx) return null

    return measureTextBounds(
      ctx,
      sample,
      layout.posX,
      layout.posY,
      layout.fontSize,
    )
  }, [fontReady, image, layout, measureCanvas, sample])

  const overflow =
    textBounds !== null &&
    (textBounds.left < 0 ||
      textBounds.top < 0 ||
      textBounds.right > imageSize.width ||
      textBounds.bottom > imageSize.height)

  const centerOffset = textBounds
    ? visualCenterOffset(textBounds, imageSize.width, imageSize.height)
    : null
  const isCentered =
    centerOffset !== null &&
    Math.abs(centerOffset.x) <= CENTER_TOLERANCE &&
    Math.abs(centerOffset.y) <= CENTER_TOLERANCE

  useEffect(() => {
    let cancelled = false

    async function loadAssets() {
      try {
        const font = new FontFace('TicketFont', `url(${fontUrl})`)
        await font.load()
        if (cancelled) return
        document.fonts.add(font)
        setFontReady(true)

        const nextImage = new Image()
        nextImage.src = ticketUrl
        await nextImage.decode()
        if (cancelled) return

        setImage(nextImage)
        setImageSize({
          width: nextImage.naturalWidth,
          height: nextImage.naturalHeight,
        })
      } catch (err) {
        if (!cancelled) {
          setAssetError(
            `Failed to load assets/ticket.png or assets/font.ttf. ${String(err)}`,
          )
        }
      }
    }

    void loadAssets()

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !image || !fontReady) return

    canvas.width = imageSize.width
    canvas.height = imageSize.height

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.drawImage(image, 0, 0)

    const centerX = imageSize.width / 2
    const centerY = imageSize.height / 2

    if (guidesVisible) {
      ctx.save()
      ctx.strokeStyle = 'rgba(255, 212, 121, 0.85)'
      ctx.lineWidth = 1.5
      ctx.setLineDash([6, 6])
      ctx.beginPath()
      ctx.moveTo(centerX, 0)
      ctx.lineTo(centerX, imageSize.height)
      ctx.moveTo(0, centerY)
      ctx.lineTo(imageSize.width, centerY)
      ctx.stroke()
      ctx.setLineDash([])
      ctx.restore()
    }

    if (textBounds) {
      ctx.save()
      ctx.strokeStyle = '#00e5ff'
      ctx.lineWidth = Math.max(1, Math.round(canvas.width / 800))
      ctx.setLineDash([8, 6])
      ctx.strokeRect(
        textBounds.left,
        textBounds.top,
        textBounds.right - textBounds.left,
        textBounds.bottom - textBounds.top,
      )
      ctx.setLineDash([])
      ctx.fillStyle = '#00e5ff'
      ctx.fillRect(
        textBounds.right - HANDLE_SIZE / 2,
        textBounds.bottom - HANDLE_SIZE / 2,
        HANDLE_SIZE,
        HANDLE_SIZE,
      )
      ctx.restore()

      ctx.font = `${layout.fontSize}px TicketFont`
      ctx.textAlign = 'center'
      ctx.textBaseline = 'alphabetic'
      ctx.fillStyle = layout.color
      ctx.fillText(
        sample,
        layout.posX,
        layout.posY + ggBaselineOffset(layout.fontSize),
      )
    }

    if (guidesVisible) {
      const markerRadius = 6
      ctx.save()
      ctx.fillStyle = '#ffd479'
      ctx.beginPath()
      ctx.moveTo(centerX, centerY - markerRadius)
      ctx.lineTo(centerX + markerRadius, centerY)
      ctx.lineTo(centerX, centerY + markerRadius)
      ctx.lineTo(centerX - markerRadius, centerY)
      ctx.closePath()
      ctx.fill()
      ctx.restore()
    }
  }, [
    guidesVisible,
    image,
    imageSize,
    fontReady,
    layout,
    sample,
    textBounds,
  ])

  function showGuidesTemporarily() {
    setGuidesVisible(true)
    if (guideTimerRef.current !== null) {
      window.clearTimeout(guideTimerRef.current)
    }
    guideTimerRef.current = window.setTimeout(() => {
      setGuidesVisible(false)
      guideTimerRef.current = null
    }, GUIDE_VISIBLE_MS)
  }

  useEffect(() => {
    return () => {
      if (guideTimerRef.current !== null) {
        window.clearTimeout(guideTimerRef.current)
      }
    }
  }, [])

  function nativePoint(event: PointerEvent<HTMLCanvasElement>) {
    const rect = event.currentTarget.getBoundingClientRect()
    return {
      x:
        ((event.clientX - rect.left) * event.currentTarget.width) / rect.width,
      y:
        ((event.clientY - rect.top) * event.currentTarget.height) /
        rect.height,
    }
  }

  function handlePointerDown(event: PointerEvent<HTMLCanvasElement>) {
    if (!sample) return

    const ctx = event.currentTarget.getContext('2d')
    if (!ctx) return

    const point = nativePoint(event)
    const bounds = measureTextBounds(
      ctx,
      sample,
      layout.posX,
      layout.posY,
      layout.fontSize,
    )
    const handleDistance = Math.hypot(
      point.x - bounds.right,
      point.y - bounds.bottom,
    )

    if (handleDistance <= HANDLE_HIT_RADIUS) {
      const startDistance = Math.hypot(
        point.x - layout.posX,
        point.y - layout.posY,
      )
      dragRef.current = {
        mode: 'resize',
        pointerId: event.pointerId,
        startDistance: Math.max(1, startDistance),
        startFontSize: layout.fontSize,
      }
    } else if (
      point.x >= bounds.left &&
      point.x <= bounds.right &&
      point.y >= bounds.top &&
      point.y <= bounds.bottom
    ) {
      dragRef.current = {
        mode: 'move',
        pointerId: event.pointerId,
        startX: point.x,
        startY: point.y,
        startPosX: layout.posX,
        startPosY: layout.posY,
      }
    } else {
      return
    }

    event.currentTarget.setPointerCapture(event.pointerId)
    showGuidesTemporarily()
  }

  function handlePointerMove(event: PointerEvent<HTMLCanvasElement>) {
    const drag = dragRef.current
    if (!drag || drag.pointerId !== event.pointerId) return

    const point = nativePoint(event)
    showGuidesTemporarily()

    if (drag.mode === 'move') {
      setLayout((current) => ({
        ...current,
        posX: clamp(
          Math.round(drag.startPosX + point.x - drag.startX),
          0,
          imageSize.width - 1,
        ),
        posY: clamp(
          Math.round(drag.startPosY + point.y - drag.startY),
          0,
          imageSize.height - 1,
        ),
      }))
      return
    }

    const distance = Math.hypot(
      point.x - layout.posX,
      point.y - layout.posY,
    )
    const nextSize = clamp(
      Math.round((drag.startFontSize * distance) / drag.startDistance),
      1,
      MAX_FONT_SIZE,
    )
    setLayout((current) => ({ ...current, fontSize: nextSize }))
  }

  function handlePointerUp(event: PointerEvent<HTMLCanvasElement>) {
    const drag = dragRef.current
    if (!drag || drag.pointerId !== event.pointerId) return

    dragRef.current = null
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId)
    }
  }

  function updatePosX(value: string) {
    if (value.trim() === '') return
    const next = Number(value)
    if (!Number.isInteger(next)) return
    setLayout((current) => ({
      ...current,
      posX: clamp(next, 0, imageSize.width - 1),
    }))
    showGuidesTemporarily()
  }

  function updatePosY(value: string) {
    if (value.trim() === '') return
    const next = Number(value)
    if (!Number.isInteger(next)) return
    setLayout((current) => ({
      ...current,
      posY: clamp(next, 0, imageSize.height - 1),
    }))
    showGuidesTemporarily()
  }

  function updateFontSize(value: string) {
    if (value.trim() === '') return
    const next = Number(value)
    if (!Number.isInteger(next)) return
    setLayout((current) => ({
      ...current,
      fontSize: clamp(next, 1, MAX_FONT_SIZE),
    }))
    showGuidesTemporarily()
  }

  function applyColor(value: string) {
    const normalized = normalizeHexColor(value)
    if (!normalized) {
      setColorError('Use a six-digit hex color such as #ffffff.')
      return
    }
    setLayout((current) => ({ ...current, color: normalized }))
    setColorDraft(normalized)
    setColorError(null)
  }

  function handleColorTextChange(value: string) {
    setColorDraft(value)
    applyColor(value)
  }

  function handleSampleChange(value: string) {
    setSample(value.replace(/\D/g, '').slice(0, 9))
    showGuidesTemporarily()
  }

  const envSnippet = buildEnvSnippet(layout)
  const ready = image !== null && fontReady

  function openResults() {
    setCopyState('idle')
    dialogRef.current?.showModal()
  }

  async function copyEnv() {
    try {
      if (!navigator.clipboard) {
        setCopyState('failed')
        return
      }
      await navigator.clipboard.writeText(envSnippet)
      setCopyState('copied')
    } catch {
      setCopyState('failed')
    }
  }

  return (
    <main className="app">
      <header className="app-header">
        <h1>Ticket number layout</h1>
        <p>
          Position the number on the ticket, then copy the values into{' '}
          <code>.env</code>.
        </p>
      </header>

      {assetError ? (
        <p className="message error" role="alert">
          {assetError}
        </p>
      ) : null}

      {!assetError && !ready ? (
        <p className="message">Loading assets...</p>
      ) : null}

      {ready ? (
        <div className="editor">
          <section className="preview">
            <canvas
              ref={canvasRef}
              className="ticket-canvas"
              width={imageSize.width}
              height={imageSize.height}
              aria-label="Ticket preview"
              onPointerDown={handlePointerDown}
              onPointerMove={handlePointerMove}
              onPointerUp={handlePointerUp}
              onPointerCancel={handlePointerUp}
            />
            {overflow ? (
              <p className="message warning" role="status">
                Text bounds exceed the ticket image. Adjust the position or size
                before exporting.
              </p>
            ) : null}
          </section>

          <aside className="controls">
            <label>
              Sample ticket number
              <input
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={9}
                value={sample}
                onChange={(event) => handleSampleChange(event.target.value)}
              />
            </label>

            <label>
              X (center)
              <input
                type="number"
                min={0}
                max={imageSize.width - 1}
                step={1}
                value={layout.posX}
                onChange={(event) => updatePosX(event.target.value)}
              />
            </label>

            <label>
              Y (center)
              <input
                type="number"
                min={0}
                max={imageSize.height - 1}
                step={1}
                value={layout.posY}
                onChange={(event) => updatePosY(event.target.value)}
              />
            </label>

            <label>
              Font size
              <input
                type="number"
                min={1}
                max={MAX_FONT_SIZE}
                step={1}
                value={layout.fontSize}
                onChange={(event) => updateFontSize(event.target.value)}
              />
            </label>

            <label>
              Font color
              <span className="color-row">
                <input
                  type="color"
                  value={layout.color}
                  onChange={(event) => applyColor(event.target.value)}
                />
                <input
                  type="text"
                  value={colorDraft}
                  spellCheck={false}
                  aria-invalid={colorError !== null}
                  onChange={(event) =>
                    handleColorTextChange(event.target.value)
                  }
                />
              </span>
            </label>
            {colorError ? (
              <p className="message error" role="alert">
                {colorError}
              </p>
            ) : null}

            {centerOffset ? (
              <p className={isCentered ? 'message centered' : 'message'}>
                {isCentered
                  ? 'Number is visually centered.'
                  : `Visual center offset: X ${formatOffset(centerOffset.x)} px, Y ${formatOffset(centerOffset.y)} px.`}
              </p>
            ) : (
              <p className="message">
                Enter a sample number to check centering.
              </p>
            )}

            <p className="hint">
              X/Y are the bot's center anchor. The offset readout uses the
              rendered glyph bounds, so a small vertical difference is normal.
              Drag the number to move it; drag the cyan corner handle to resize
              it. Center guides appear briefly after each adjustment.
            </p>
          </aside>
        </div>
      ) : null}

      <div className="actions">
        <button
          type="button"
          onClick={openResults}
          disabled={!sample || !ready}
        >
          Open .env values
        </button>
      </div>

      <dialog
        ref={dialogRef}
        className="results"
        aria-labelledby="results-title"
        onClose={() => setCopyState('idle')}
      >
        <h2 id="results-title">.env values</h2>
        <p>Paste these lines into the bot's .env file, then restart the bot.</p>
        <pre>{envSnippet}</pre>
        <div className="dialog-actions">
          <button type="button" onClick={copyEnv}>
            Copy
          </button>
          <button
            type="button"
            className="secondary"
            onClick={() => dialogRef.current?.close()}
          >
            Close
          </button>
        </div>
        {copyState === 'copied' ? <p>Copied.</p> : null}
        {copyState === 'failed' ? (
          <p className="message error">
            Copy failed. Select the text above and copy it manually.
          </p>
        ) : null}
      </dialog>
    </main>
  )
}

export default App
