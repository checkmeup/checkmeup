import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getMock, postMock, patchMock, deleteMock } = vi.hoisted(() => ({
  getMock: vi.fn(),
  postMock: vi.fn(),
  patchMock: vi.fn(),
  deleteMock: vi.fn(),
}))

vi.mock('./client', () => ({
  api: { get: getMock, post: postMock, patch: patchMock, delete: deleteMock },
}))

import { monitorsApi } from './monitors'

beforeEach(() => {
  getMock.mockReset()
  postMock.mockReset()
  patchMock.mockReset()
  deleteMock.mockReset()
})

const cronMonitor = {
  id: 'c1',
  name: 'Nightly backup',
  schedule: '0 2 * * *',
  gracePeriodMins: 5,
  pingToken: 'tok',
  pingUrl: 'https://checkmeup.net/ping/tok',
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  lastPingAt: '2026-06-21T02:00:00Z',
  nextPingAt: '2026-06-22T02:00:00Z',
  createdAt: '2026-06-01T00:00:00Z',
}

const createCronInput = {
  name: 'Nightly backup',
  schedule: '0 2 * * *',
  gracePeriodMins: 5,
  alertAfterNFailures: 1,
  maxAlertsPerIncident: 3,
}

const updateCronInput = {
  name: 'Nightly backup',
  schedule: '0 2 * * *',
  gracePeriodMins: 5,
  alertsEnabled: true,
  alertAfterNFailures: 1,
  maxAlertsPerIncident: 3,
  channelIds: ['ch1'],
}

describe('monitorsApi cron', () => {
  it('listCron fetches the cron monitor list', async () => {
    getMock.mockResolvedValueOnce([cronMonitor])

    const result = await monitorsApi.listCron()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/')
    expect(result).toEqual([cronMonitor])
  })

  it('getCron fetches a single cron monitor detail by id', async () => {
    const detail = { monitor: cronMonitor, pings: [], incidents: [] }
    getMock.mockResolvedValueOnce(detail)

    const result = await monitorsApi.getCron('c1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/c1/')
    expect(result).toEqual(detail)
  })

  it('createCron posts the input to create a cron monitor', async () => {
    postMock.mockResolvedValueOnce(cronMonitor)

    const result = await monitorsApi.createCron(createCronInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/', createCronInput)
    expect(result).toEqual(cronMonitor)
  })

  it('updateCron patches the input to update a cron monitor by id', async () => {
    patchMock.mockResolvedValueOnce(cronMonitor)

    const result = await monitorsApi.updateCron('c1', updateCronInput)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/c1/', updateCronInput)
    expect(result).toEqual(cronMonitor)
  })

  it('pauseCron posts to the pause endpoint with no body', async () => {
    const paused = { ...cronMonitor, status: 'paused' as const }
    postMock.mockResolvedValueOnce(paused)

    const result = await monitorsApi.pauseCron('c1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/c1/pause', {})
    expect(result).toEqual(paused)
  })

  it('resumeCron posts to the resume endpoint with no body', async () => {
    postMock.mockResolvedValueOnce(cronMonitor)

    const result = await monitorsApi.resumeCron('c1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/c1/resume', {})
    expect(result).toEqual(cronMonitor)
  })

  it('deleteCron deletes a cron monitor by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await monitorsApi.deleteCron('c1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/c1/')
  })

  it('getCronPings fetches the first page of pings by default', async () => {
    const pings = [{ id: 'p1', receivedAt: '2026-06-22T02:00:00Z', sourceIp: '1.2.3.4' }]
    getMock.mockResolvedValueOnce(pings)

    const result = await monitorsApi.getCronPings('c1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/c1/pings?page=1')
    expect(result).toEqual(pings)
  })

  it('getCronPings fetches an explicit page of pings', async () => {
    getMock.mockResolvedValueOnce([])

    await monitorsApi.getCronPings('c1', 3)

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/cron/c1/pings?page=3')
  })
})

const sslMonitor = {
  id: 's1',
  name: 'Primary cert',
  hostname: 'checkmeup.net',
  status: 'up' as const,
  alertsEnabled: true,
  expiresAt: '2026-12-01T00:00:00Z',
  issuer: "Let's Encrypt",
  errorMsg: null,
  daysUntilExpiry: 162,
  lastCheckedAt: '2026-06-22T00:00:00Z',
  createdAt: '2026-01-01T00:00:00Z',
}

const createSSLInput = { name: 'Primary cert', hostname: 'checkmeup.net' }

const updateSSLInput = {
  name: 'Primary cert',
  hostname: 'checkmeup.net',
  alertsEnabled: true,
  alertAfterNFailures: 1,
  maxAlertsPerIncident: 3,
  channelIds: ['ch1'],
}

describe('monitorsApi ssl', () => {
  it('listSSL fetches the SSL monitor list', async () => {
    getMock.mockResolvedValueOnce([sslMonitor])

    const result = await monitorsApi.listSSL()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/ssl/')
    expect(result).toEqual([sslMonitor])
  })

  it('getSSL fetches a single SSL monitor by id', async () => {
    getMock.mockResolvedValueOnce(sslMonitor)

    const result = await monitorsApi.getSSL('s1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/ssl/s1/')
    expect(result).toEqual(sslMonitor)
  })

  it('createSSL posts the input to create an SSL monitor', async () => {
    postMock.mockResolvedValueOnce(sslMonitor)

    const result = await monitorsApi.createSSL(createSSLInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/ssl/', createSSLInput)
    expect(result).toEqual(sslMonitor)
  })

  it('updateSSL patches the input to update an SSL monitor by id', async () => {
    patchMock.mockResolvedValueOnce(sslMonitor)

    const result = await monitorsApi.updateSSL('s1', updateSSLInput)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/ssl/s1/', updateSSLInput)
    expect(result).toEqual(sslMonitor)
  })

  it('pauseSSL posts to the pause endpoint with no body', async () => {
    const paused = { ...sslMonitor, status: 'paused' as const }
    postMock.mockResolvedValueOnce(paused)

    const result = await monitorsApi.pauseSSL('s1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/ssl/s1/pause', {})
    expect(result).toEqual(paused)
  })

  it('resumeSSL posts to the resume endpoint with no body', async () => {
    postMock.mockResolvedValueOnce(sslMonitor)

    const result = await monitorsApi.resumeSSL('s1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/ssl/s1/resume', {})
    expect(result).toEqual(sslMonitor)
  })

  it('deleteSSL deletes an SSL monitor by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await monitorsApi.deleteSSL('s1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/ssl/s1/')
  })
})

const domainMonitor = {
  id: 'd1',
  name: 'Primary domain',
  domain: 'checkmeup.net',
  status: 'up' as const,
  alertsEnabled: true,
  expiresAt: '2027-03-01T00:00:00Z',
  registrar: 'Example Registrar, LLC',
  errorMsg: null,
  daysUntilExpiry: 250,
  lastCheckedAt: '2026-06-22T00:00:00Z',
  createdAt: '2026-01-01T00:00:00Z',
}

const createDomainInput = { name: 'Primary domain', domain: 'checkmeup.net' }

const updateDomainInput = {
  name: 'Primary domain',
  domain: 'checkmeup.net',
  alertsEnabled: true,
  alertAfterNFailures: 1,
  maxAlertsPerIncident: 3,
  channelIds: ['ch1'],
}

describe('monitorsApi domain', () => {
  it('listDomain fetches the domain monitor list', async () => {
    getMock.mockResolvedValueOnce([domainMonitor])

    const result = await monitorsApi.listDomain()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/domain/')
    expect(result).toEqual([domainMonitor])
  })

  it('getDomain fetches a single domain monitor by id', async () => {
    getMock.mockResolvedValueOnce(domainMonitor)

    const result = await monitorsApi.getDomain('d1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/domain/d1/')
    expect(result).toEqual(domainMonitor)
  })

  it('createDomain posts the input to create a domain monitor', async () => {
    postMock.mockResolvedValueOnce(domainMonitor)

    const result = await monitorsApi.createDomain(createDomainInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/domain/', createDomainInput)
    expect(result).toEqual(domainMonitor)
  })

  it('updateDomain patches the input to update a domain monitor by id', async () => {
    patchMock.mockResolvedValueOnce(domainMonitor)

    const result = await monitorsApi.updateDomain('d1', updateDomainInput)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/monitors/domain/d1/',
      updateDomainInput,
    )
    expect(result).toEqual(domainMonitor)
  })

  it('pauseDomain posts to the pause endpoint with no body', async () => {
    const paused = { ...domainMonitor, status: 'paused' as const }
    postMock.mockResolvedValueOnce(paused)

    const result = await monitorsApi.pauseDomain('d1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/domain/d1/pause', {})
    expect(result).toEqual(paused)
  })

  it('resumeDomain posts to the resume endpoint with no body', async () => {
    postMock.mockResolvedValueOnce(domainMonitor)

    const result = await monitorsApi.resumeDomain('d1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/domain/d1/resume', {})
    expect(result).toEqual(domainMonitor)
  })

  it('deleteDomain deletes a domain monitor by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await monitorsApi.deleteDomain('d1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/domain/d1/')
  })
})

const uptimeMonitor = {
  id: 'u1',
  name: 'Marketing site',
  url: 'https://checkmeup.net',
  intervalMins: 5,
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  lastCheckedAt: '2026-06-22T00:00:00Z',
  createdAt: '2026-01-01T00:00:00Z',
  uptime24h: 99.9,
  keyword: null,
  keywordMode: 'contains' as const,
  keywordCaseSensitive: false,
  jsonAssertions: [],
  maxResponseTimeMs: 10000,
  httpMethod: 'GET' as const,
  acceptedStatusCodes: [200],
}

const createUptimeInput = {
  name: 'Marketing site',
  url: 'https://checkmeup.net',
  intervalMins: 5,
  alertAfterNFailures: 1,
  maxAlertsPerIncident: 3,
  jsonAssertions: [],
  maxResponseTimeMs: 10000,
  httpMethod: 'GET' as const,
  acceptedStatusCodes: [200],
  keyword: '',
  keywordMode: 'contains' as const,
  keywordCaseSensitive: false,
}

const updateUptimeInput = {
  name: 'Marketing site',
  url: 'https://checkmeup.net',
  intervalMins: 5,
  alertsEnabled: true,
  alertAfterNFailures: 1,
  maxAlertsPerIncident: 3,
  jsonAssertions: [],
  maxResponseTimeMs: 10000,
  httpMethod: 'GET' as const,
  acceptedStatusCodes: [200],
  keyword: '',
  keywordMode: 'contains' as const,
  keywordCaseSensitive: false,
  channelIds: ['ch1'],
}

describe('monitorsApi uptime', () => {
  it('listUptime fetches the uptime monitor list', async () => {
    getMock.mockResolvedValueOnce([uptimeMonitor])

    const result = await monitorsApi.listUptime()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/uptime/')
    expect(result).toEqual([uptimeMonitor])
  })

  it('getUptime fetches a single uptime monitor detail by id', async () => {
    const detail = {
      monitor: uptimeMonitor,
      chartData: [],
      checks: [],
      incidents: [],
      stats: { uptime24h: 99.9, uptime7d: 99.8, uptime30d: 99.7 },
    }
    getMock.mockResolvedValueOnce(detail)

    const result = await monitorsApi.getUptime('u1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/uptime/u1/')
    expect(result).toEqual(detail)
  })

  it('createUptime posts the input to create an uptime monitor', async () => {
    postMock.mockResolvedValueOnce(uptimeMonitor)

    const result = await monitorsApi.createUptime(createUptimeInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/uptime/', createUptimeInput)
    expect(result).toEqual(uptimeMonitor)
  })

  it('updateUptime patches the input to update an uptime monitor by id', async () => {
    patchMock.mockResolvedValueOnce(uptimeMonitor)

    const result = await monitorsApi.updateUptime('u1', updateUptimeInput)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/monitors/uptime/u1/',
      updateUptimeInput,
    )
    expect(result).toEqual(uptimeMonitor)
  })

  it('pauseUptime posts to the pause endpoint with no body', async () => {
    const paused = { ...uptimeMonitor, status: 'paused' as const }
    postMock.mockResolvedValueOnce(paused)

    const result = await monitorsApi.pauseUptime('u1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/uptime/u1/pause', {})
    expect(result).toEqual(paused)
  })

  it('resumeUptime posts to the resume endpoint with no body', async () => {
    postMock.mockResolvedValueOnce(uptimeMonitor)

    const result = await monitorsApi.resumeUptime('u1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/uptime/u1/resume', {})
    expect(result).toEqual(uptimeMonitor)
  })

  it('deleteUptime deletes an uptime monitor by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await monitorsApi.deleteUptime('u1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/uptime/u1/')
  })
})

const portMonitor = {
  id: 'p1',
  name: 'SMTP',
  host: 'mail.checkmeup.net',
  port: 25,
  expectedState: 'open' as const,
  intervalMins: 5,
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  alertAfterNFailures: 0,
  lastCheckedAt: '2026-06-22T00:00:00Z',
  createdAt: '2026-01-01T00:00:00Z',
  uptime24h: 99.9,
}

const createPortInput = {
  name: 'SMTP',
  host: 'mail.checkmeup.net',
  port: 25,
  expectedState: 'open' as const,
  intervalMins: 5,
  maxAlertsPerIncident: 3,
  alertAfterNFailures: 0,
}

const updatePortInput = {
  name: 'SMTP',
  host: 'mail.checkmeup.net',
  port: 25,
  expectedState: 'open' as const,
  intervalMins: 5,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  alertAfterNFailures: 0,
  channelIds: ['ch1'],
}

describe('monitorsApi port', () => {
  it('listPort fetches the port monitor list', async () => {
    getMock.mockResolvedValueOnce([portMonitor])

    const result = await monitorsApi.listPort()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/port/')
    expect(result).toEqual([portMonitor])
  })

  it('getPort fetches a single port monitor detail by id', async () => {
    getMock.mockResolvedValueOnce(portMonitor)

    const result = await monitorsApi.getPort('p1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/port/p1/')
    expect(result).toEqual(portMonitor)
  })

  it('createPort posts the input to create a port monitor', async () => {
    postMock.mockResolvedValueOnce(portMonitor)

    const result = await monitorsApi.createPort(createPortInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/port/', createPortInput)
    expect(result).toEqual(portMonitor)
  })

  it('updatePort patches the input to update a port monitor by id', async () => {
    patchMock.mockResolvedValueOnce(portMonitor)

    const result = await monitorsApi.updatePort('p1', updatePortInput)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/port/p1/', updatePortInput)
    expect(result).toEqual(portMonitor)
  })

  it('pausePort posts to the pause endpoint with no body', async () => {
    const paused = { ...portMonitor, status: 'paused' as const }
    postMock.mockResolvedValueOnce(paused)

    const result = await monitorsApi.pausePort('p1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/port/p1/pause', {})
    expect(result).toEqual(paused)
  })

  it('resumePort posts to the resume endpoint with no body', async () => {
    postMock.mockResolvedValueOnce(portMonitor)

    const result = await monitorsApi.resumePort('p1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/port/p1/resume', {})
    expect(result).toEqual(portMonitor)
  })

  it('deletePort deletes a port monitor by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await monitorsApi.deletePort('p1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/port/p1/')
  })
})

const dnsMonitor = {
  id: 'd1',
  name: 'Apex A record',
  hostname: 'checkmeup.net',
  recordType: 'A' as const,
  expectedValue: '1.2.3.4',
  baselineCaptured: false,
  lastResolvedValue: '1.2.3.4',
  intervalMins: 5,
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  alertAfterNFailures: 0,
  lastCheckedAt: '2026-06-22T00:00:00Z',
  createdAt: '2026-01-01T00:00:00Z',
  uptime24h: 99.9,
}

const createDNSInput = {
  name: 'Apex A record',
  hostname: 'checkmeup.net',
  recordType: 'A' as const,
  expectedValue: '1.2.3.4',
  intervalMins: 5,
  maxAlertsPerIncident: 3,
  alertAfterNFailures: 0,
}

const updateDNSInput = {
  name: 'Apex A record',
  hostname: 'checkmeup.net',
  recordType: 'A' as const,
  expectedValue: '1.2.3.4',
  intervalMins: 5,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  alertAfterNFailures: 0,
  channelIds: ['ch1'],
}

describe('monitorsApi dns', () => {
  it('listDns fetches the dns monitor list', async () => {
    getMock.mockResolvedValueOnce([dnsMonitor])

    const result = await monitorsApi.listDns()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/dns/')
    expect(result).toEqual([dnsMonitor])
  })

  it('getDns fetches a single dns monitor detail by id', async () => {
    getMock.mockResolvedValueOnce(dnsMonitor)

    const result = await monitorsApi.getDns('d1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/dns/d1/')
    expect(result).toEqual(dnsMonitor)
  })

  it('createDns posts the input to create a dns monitor', async () => {
    postMock.mockResolvedValueOnce(dnsMonitor)

    const result = await monitorsApi.createDns(createDNSInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/dns/', createDNSInput)
    expect(result).toEqual(dnsMonitor)
  })

  it('updateDns patches the input to update a dns monitor by id', async () => {
    patchMock.mockResolvedValueOnce(dnsMonitor)

    const result = await monitorsApi.updateDns('d1', updateDNSInput)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/dns/d1/', updateDNSInput)
    expect(result).toEqual(dnsMonitor)
  })

  it('pauseDns posts to the pause endpoint with no body', async () => {
    const paused = { ...dnsMonitor, status: 'paused' as const }
    postMock.mockResolvedValueOnce(paused)

    const result = await monitorsApi.pauseDns('d1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/dns/d1/pause', {})
    expect(result).toEqual(paused)
  })

  it('resumeDns posts to the resume endpoint with no body', async () => {
    postMock.mockResolvedValueOnce(dnsMonitor)

    const result = await monitorsApi.resumeDns('d1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/dns/d1/resume', {})
    expect(result).toEqual(dnsMonitor)
  })

  it('deleteDns deletes a dns monitor by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await monitorsApi.deleteDns('d1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/monitors/dns/d1/')
  })
})
