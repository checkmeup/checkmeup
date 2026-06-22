import { beforeEach, describe, expect, it, vi } from 'vitest'

const { postMock } = vi.hoisted(() => ({
  postMock: vi.fn(),
}))

vi.mock('./client', () => ({
  api: { post: postMock },
}))

import { suggestionsApi } from './suggestions'

beforeEach(() => {
  postMock.mockReset()
})

describe('suggestionsApi.submit', () => {
  it('posts the text wrapped in an object to the suggestions endpoint', async () => {
    postMock.mockResolvedValueOnce(undefined)

    await suggestionsApi.submit('Please add dark mode')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/suggestions', {
      text: 'Please add dark mode',
    })
  })
})
