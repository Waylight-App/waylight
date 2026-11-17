import type { IntentSpec } from './intent';

/**
 * The intents of the first sources (#156). Terms are folded (no umlauts,
 * lower case); weights say how strongly a word points at the intent.
 * Adding a question means adding words here – the regression table in
 * intents.test.ts shows the effect.
 */
const t = (weight: number, ...terms: string[]) => terms.map((term) => ({ term, weight }));

export const INTENTS: IntentSpec[] = [
  {
    id: 'system.cpu',
    help: {
      group: 'home',
      de: 'Wie stark der Prozessor deines Servers gerade arbeitet',
      en: 'How hard the processor of your server is working',
      examples: { de: ['cpu', 'prozessor auslastung'], en: ['cpu', 'processor usage'] },
    },
    source: 'system',
    prefixes: ['cpu'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [
      ...t(3, 'cpu', 'prozessor', 'processor'),
      ...t(2, 'auslastung', 'ausgelastet', 'usage', 'utilization', 'busy'),
    ],
  },
  {
    id: 'system.memory',
    help: {
      group: 'home',
      de: 'Wie viel Arbeitsspeicher belegt und frei ist',
      en: 'How much memory is in use and free',
      examples: { de: ['ram', 'arbeitsspeicher'], en: ['ram', 'memory free'] },
    },
    source: 'system',
    prefixes: ['ram'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [...t(3, 'ram', 'arbeitsspeicher', 'memory', 'mem'), ...t(1, 'speicher', 'frei', 'free')],
  },
  {
    id: 'system.disk',
    help: {
      group: 'home',
      de: 'Wie voll die Festplatten sind',
      en: 'How full the disks are',
      examples: { de: ['festplatte voll', 'speicherplatz'], en: ['disk space', 'storage full'] },
    },
    source: 'system',
    prefixes: ['disk'],
    answer: 'list',
    widgetKind: 'system',
    terms: [
      ...t(3, 'disk', 'disks', 'festplatte', 'platte', 'speicherplatz', 'storage', 'laufwerk'),
      ...t(1.5, 'voll', 'full', 'space', 'platz'),
      ...t(1, 'frei', 'free', 'speicher'),
    ],
  },
  {
    id: 'system.load',
    help: {
      group: 'home',
      de: 'Die Systemlast des Servers',
      en: 'The load of the server',
      examples: { de: ['systemlast'], en: ['load'] },
    },
    source: 'system',
    prefixes: ['load'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [...t(3, 'load', 'loadavg', 'systemlast'), ...t(1.5, 'last')],
  },
  {
    id: 'system.uptime',
    help: {
      group: 'home',
      de: 'Wie lange der Server schon läuft',
      en: 'How long the server has been running',
      examples: { de: ['laufzeit', 'uptime'], en: ['uptime', 'reboot since'] },
    },
    source: 'system',
    prefixes: ['uptime'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [
      ...t(3, 'uptime', 'laufzeit', 'betriebszeit'),
      ...t(1.5, 'neustart', 'reboot', 'seit', 'since', 'lange', 'long'),
      ...t(1, 'lauft', 'laeuft', 'running', 'server'),
    ],
  },
  {
    id: 'system.temperature',
    help: {
      group: 'home',
      de: 'Wie warm der Server ist',
      en: 'How warm the server is',
      examples: { de: ['temperatur server'], en: ['server temperature'] },
    },
    source: 'system',
    answer: 'metric',
    widgetKind: 'system',
    terms: [
      ...t(1.5, 'temperatur', 'temperature', 'temp'),
      ...t(1, 'heiss', 'hot', 'server'),
      // "cpu temperatur" asks for the temperature, not the load.
      ...t(2, 'cpu'),
    ],
  },
  {
    id: 'docker.overview',
    help: {
      group: 'home',
      de: 'Welche Docker-Container laufen und welche gestoppt sind',
      en: 'Which Docker containers are running or stopped',
      examples: { de: ['docker', 'container gestoppt'], en: ['docker', 'containers running'] },
    },
    source: 'docker',
    answer: 'list',
    widgetKind: 'docker',
    terms: [
      ...t(3, 'docker', 'container', 'containers'),
      ...t(1, 'laufen', 'laufende', 'running', 'gestoppt', 'stopped', 'down', 'kaputt'),
    ],
  },
  {
    id: 'docker.status',
    help: {
      group: 'home',
      de: 'Der Zustand eines einzelnen Containers',
      en: 'The state of one container',
      examples: { de: ['läuft jellyfin'], en: ['jellyfin status'] },
    },
    source: 'docker',
    prefixes: ['docker'],
    answer: 'status',
    widgetKind: 'docker',
    slots: [{ name: 'container', type: 'entity', entityKind: 'container', required: true }],
    terms: t(1, 'status', 'lauft', 'laeuft', 'running', 'image', 'update'),
  },
  {
    id: 'docker.restart',
    help: {
      group: 'home',
      de: 'Einen Container neu starten, stoppen oder starten (mit Rückfrage)',
      en: 'Restart, stop or start a container (asks first)',
      examples: { de: ['jellyfin neu starten'], en: ['restart jellyfin'] },
    },
    source: 'docker',
    prefixes: ['docker'],
    answer: 'actions',
    widgetKind: 'docker',
    slots: [
      { name: 'action', type: 'action', required: true },
      { name: 'container', type: 'entity', entityKind: 'container', required: true },
    ],
    terms: [],
  },
  {
    id: 'services.overview',
    help: {
      group: 'home',
      de: 'Welche Dienste erreichbar sind und welche nicht',
      en: 'Which services are reachable and which are down',
      examples: { de: ['dienste offline'], en: ['services down'] },
    },
    source: 'services',
    answer: 'list',
    widgetKind: 'services',
    terms: [
      ...t(3, 'dienste', 'services', 'dienst'),
      ...t(2, 'offline', 'down', 'erreichbar', 'reachable', 'ausgefallen', 'alles', 'everything'),
      ...t(1, 'online', 'up', 'ok'),
    ],
  },
  {
    id: 'services.status',
    help: {
      group: 'home',
      de: 'Ob ein bestimmter Dienst erreichbar ist',
      en: 'Whether one service is reachable',
      examples: { de: ['nas erreichbar'], en: ['nas online'] },
    },
    source: 'services',
    prefixes: ['service'],
    answer: 'status',
    widgetKind: 'services',
    slots: [{ name: 'service', type: 'entity', entityKind: 'service', required: true }],
    terms: t(1, 'status', 'online', 'erreichbar', 'antwortzeit', 'latency', 'ping'),
  },
  {
    id: 'weather.forecast',
    help: {
      group: 'day',
      de: 'Das Wetter heute oder an einem anderen Tag',
      en: 'The weather today or on another day',
      examples: { de: ['wetter morgen', 'regnet es'], en: ['weather tomorrow', 'rain'] },
    },
    source: 'weather',
    prefixes: ['wetter', 'weather'],
    answer: 'metric',
    widgetKind: 'weather',
    slots: [
      { name: 'place', type: 'place' },
      { name: 'day', type: 'day' },
    ],
    terms: [
      ...t(3, 'wetter', 'weather', 'vorhersage', 'forecast'),
      ...t(
        2,
        'regen',
        'regnet',
        'rain',
        'raining',
        'schnee',
        'schneit',
        'snow',
        'snowing',
        'sonne',
        'sonnig',
        'sunny',
      ),
      ...t(2, 'gewitter', 'thunderstorm', 'sturm', 'storm', 'wind', 'windig', 'windy', 'nebel', 'fog'),
      ...t(1.5, 'temperatur', 'temperature', 'draussen', 'outside'),
      ...t(2, 'schirm', 'regenschirm', 'umbrella'),
      ...t(1, 'kalt', 'cold', 'warm', 'jacke', 'jacket', 'mantel', 'coat', 'sonnencreme', 'sunscreen'),
    ],
  },
];
