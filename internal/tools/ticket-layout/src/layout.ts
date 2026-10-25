export type LayoutState = {
  posX: number
  posY: number
  fontSize: number
  color: string
}

export const DEFAULT_LAYOUT: LayoutState = {
  posX: 250,
  posY: 250,
  fontSize: 100,
  color: '#ffffff',
}

export const MAX_FONT_SIZE = 500
export const HANDLE_SIZE = 12
export const HANDLE_HIT_RADIUS = 16

export type TextBounds = {
  left: number
  right: number
  top: number
  bottom: number
}

export type CenterOffset = {
  x: number
  y: number
}

export const CENTER_TOLERANCE = 1

export function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

export function ggBaselineOffset(fontSize: number): number {
  return (fontSize * 72) / 96 / 2
}

export function normalizeHexColor(value: string): string | null {
  const trimmed = value.trim()
  const withHash = trimmed.startsWith('#') ? trimmed : `#${trimmed}`
  return /^#[0-9a-fA-F]{6}$/.test(withHash) ? withHash.toLowerCase() : null
}

export function buildEnvSnippet(state: LayoutState): string {
  return [
    `TICKET_NUMBER_POS_X=${state.posX}`,
    `TICKET_NUMBER_POS_Y=${state.posY}`,
    `TICKET_NUMBER_FONT_SIZE=${state.fontSize}`,
    `TICKET_NUMBER_COLOR=${state.color}`,
  ].join('\n')
}

export function visualCenterOffset(
  bounds: TextBounds,
  width: number,
  height: number,
): CenterOffset {
  return {
    x: (bounds.left + bounds.right) / 2 - width / 2,
    y: (bounds.top + bounds.bottom) / 2 - height / 2,
  }
}

export function formatOffset(value: number): string {
  const rounded = Math.round(value)
  return rounded > 0 ? `+${rounded}` : `${rounded}`
}

export function measureTextBounds(
  ctx: CanvasRenderingContext2D,
  text: string,
  posX: number,
  posY: number,
  fontSize: number,
): TextBounds {
  ctx.font = `${fontSize}px TicketFont`
  const metrics = ctx.measureText(text)
  const baselineY = posY + ggBaselineOffset(fontSize)
  const ascent = metrics.actualBoundingBoxAscent || fontSize * 0.75
  const descent = metrics.actualBoundingBoxDescent || fontSize * 0.25

  return {
    left: posX - metrics.width / 2,
    right: posX + metrics.width / 2,
    top: baselineY - ascent,
    bottom: baselineY + descent,
  }
}
