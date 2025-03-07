import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { filesystem, print, prompt, system } from 'gluegun';
import { Args, Command, Flags } from '@oclif/core';
import { appendApiVersionForGraph } from '../command-helpers/compiler.js';
import { ContractService } from '../command-helpers/contracts.js';
import { resolveFile } from '../command-helpers/file-resolver.js';
import { DEFAULT_IPFS_URL } from '../command-helpers/ipfs.js';
import { initNetworksConfig } from '../command-helpers/network.js';
import { chooseNodeUrl } from '../command-helpers/node.js';
import { PromptManager } from '../command-helpers/prompt-manager.js';
import { loadRegistry } from '../command-helpers/registry.js';
import { retryWithPrompt } from '../command-helpers/retry.js';
import { generateScaffold, writeScaffold } from '../command-helpers/scaffold.js';
import { sortWithPriority } from '../command-helpers/sort.js';
import { withSpinner } from '../command-helpers/spinner.js';
import { getSubgraphBasename } from '../command-helpers/subgraph.js';
import { GRAPH_CLI_SHARED_HEADERS } from '../constants.js';
import debugFactory from '../debug.js';
import Protocol from '../protocols/index.js';
import { abiEvents } from '../scaffold/schema.js';
import Schema from '../schema.js';
import { createIpfsClient, loadSubgraphSchemaFromIPFS } from '../utils.js';
import { validateContract } from '../validation/index.js';
import AddCommand from './add.js';
const protocolChoices = Array.from(Protocol.availableProtocols().keys());
const initDebugger = debugFactory('graph-cli:commands:init');
const DEFAULT_EXAMPLE_SUBGRAPH = 'ethereum-gravatar';
const DEFAULT_CONTRACT_NAME = 'Contract';
export default class InitCommand extends Command {
    static description = 'Creates a new subgraph with basic scaffolding.';
    static args = {
        subgraphName: Args.string(),
        directory: Args.string(),
    };
    static flags = {
        help: Flags.help({
            char: 'h',
        }),
        protocol: Flags.string({
            options: protocolChoices,
        }),
        node: Flags.string({
            summary: 'Graph node for which to initialize.',
            char: 'g',
        }),
        'from-contract': Flags.string({
            description: 'Creates a scaffold based on an existing contract.',
            exclusive: ['from-example'],
        }),
        'from-example': Flags.string({
            description: 'Creates a scaffold based on an example subgraph.',
            // TODO: using a default sets the value and therefore requires not to have --from-contract
            // default: 'Contract',
            exclusive: ['from-contract', 'spkg'],
        }),
        'contract-name': Flags.string({
            helpGroup: 'Scaffold from contract',
            description: 'Name of the contract.',
            dependsOn: ['from-contract'],
        }),
        'index-events': Flags.boolean({
            helpGroup: 'Scaffold from contract',
            description: 'Index contract events as entities.',
            dependsOn: ['from-contract'],
        }),
        'skip-install': Flags.boolean({
            summary: 'Skip installing dependencies.',
            default: false,
        }),
        'skip-git': Flags.boolean({
            summary: 'Skip initializing a Git repository.',
            default: false,
        }),
        'start-block': Flags.string({
            helpGroup: 'Scaffold from contract',
            description: 'Block number to start indexing from.',
            // TODO: using a default sets the value and therefore requires --from-contract
            // default: '0',
            dependsOn: ['from-contract'],
        }),
        abi: Flags.string({
            summary: 'Path to the contract ABI',
            // TODO: using a default sets the value and therefore requires --from-contract
            // default: '*Download from Etherscan*',
            dependsOn: ['from-contract'],
        }),
        spkg: Flags.string({
            summary: 'Path to the SPKG file',
        }),
        network: Flags.string({
            summary: 'Network the contract is deployed to.',
            description: 'Refer to https://github.com/graphprotocol/networks-registry/ for supported networks',
        }),
        ipfs: Flags.string({
            summary: 'IPFS node to use for fetching subgraph data.',
            char: 'i',
            default: DEFAULT_IPFS_URL,
            hidden: true,
        }),
    };
    async run() {
        const { args: { subgraphName, directory }, flags, } = await this.parse(InitCommand);
        const { protocol, node: nodeFlag, 'from-contract': fromContract, 'contract-name': contractName, 'from-example': fromExample, 'index-events': indexEvents, 'skip-install': skipInstall, 'skip-git': skipGit, ipfs, network, abi: abiPath, 'start-block': startBlock, spkg: spkgPath, } = flags;
        initDebugger('Flags: %O', flags);
        // if (skipGit) {
        //     this.warn('The --skip-git flag will be removed in the next major version. By default we will stop initializing a Git repository.');
        // }
        if ((fromContract || spkgPath) && !network && !fromExample) {
            this.error('--network is required when using --from-contract or --spkg');
        }
        const { node } = chooseNodeUrl({
            node: nodeFlag,
        });
        // Detect git
        const git = system.which('git');
        if (!git) {
            this.error('Git was not found on your system. Please install "git" so it is in $PATH.', {
                exit: 1,
            });
        }
        // Detect Yarn and/or NPM
        const yarn = system.which('yarn');
        const npm = system.which('npm');
        if (!yarn && !npm) {
            this.error(`Neither Yarn nor NPM were found on your system. Please install one of them.`, {
                exit: 1,
            });
        }
        const commands = {
            link: yarn ? 'yarn link @graphprotocol/graph-cli' : 'npm link @graphprotocol/graph-cli',
            install: yarn ? 'yarn' : 'npm install',
            codegen: yarn ? 'yarn codegen' : 'npm run codegen',
            deploy: yarn ? 'yarn deploy' : 'npm run deploy',
        };
        // If all parameters are provided from the command-line,
        // go straight to creating the subgraph from the example
        if (fromExample && subgraphName && directory) {
            await initSubgraphFromExample.bind(this)({
                fromExample,
                directory,
                subgraphName,
                skipInstall,
                skipGit,
            }, { commands });
            // Exit with success
            return this.exit(0);
        }
        // Will be assigned below if ethereum
        let abi;
        // If all parameters are provided from the command-line,
        // go straight to creating the subgraph from an existing contract
        if ((fromContract || spkgPath) && protocol && subgraphName && directory && network && node) {
            const registry = await loadRegistry();
            const contractService = new ContractService(registry);
            if (!protocolChoices.includes(protocol)) {
                this.error(`Protocol '${protocol}' is not supported, choose from these options: ${protocolChoices.join(', ')}`, { exit: 1 });
            }
            const protocolInstance = new Protocol(protocol);
            if (protocolInstance.hasABIs()) {
                const ABI = protocolInstance.getABI();
                if (abiPath) {
                    try {
                        abi = loadAbiFromFile(ABI, abiPath);
                    }
                    catch (e) {
                        this.error(`Failed to load ABI: ${e.message}`, { exit: 1 });
                    }
                }
                else {
                    try {
                        abi = await contractService.getABI(ABI, network, fromContract);
                    }
                    catch (e) {
                        this.exit(1);
                    }
                }
            }
            await initSubgraphFromContract.bind(this)({
                protocolInstance,
                abi,
                directory,
                source: fromContract,
                indexEvents,
                network,
                subgraphName,
                contractName: contractName || DEFAULT_CONTRACT_NAME,
                node,
                startBlock,
                spkgPath,
                skipInstall,
                skipGit,
                ipfsUrl: ipfs,
            }, { commands, addContract: false });
            // Exit with success
            return this.exit(0);
        }
        if (fromExample) {
            const answers = await processFromExampleInitForm.bind(this)({
                subgraphName,
                directory,
            });
            if (!answers) {
                this.exit(1);
            }
            await initSubgraphFromExample.bind(this)({
                fromExample,
                subgraphName: answers.subgraphName,
                directory: answers.directory,
                skipInstall,
                skipGit,
            }, { commands });
        }
        else {
            // Otherwise, take the user through the interactive form
            const answers = await processInitForm.bind(this)({
                network,
                abi,
                abiPath,
                directory,
                source: fromContract,
                indexEvents,
                fromExample,
                subgraphName,
                contractName,
                startBlock,
                spkgPath,
                ipfsUrl: ipfs,
            });
            if (!answers) {
                this.exit(1);
            }
            await initSubgraphFromContract.bind(this)({
                protocolInstance: answers.protocolInstance,
                subgraphName: answers.subgraphName,
                directory: answers.directory,
                abi: answers.abi,
                network: network,
                source: answers.source,
                indexEvents: answers.indexEvents,
                contractName: answers.contractName || DEFAULT_CONTRACT_NAME,
                node,
                startBlock: answers.startBlock,
                spkgPath: answers.spkgPath,
                skipInstall,
                skipGit,
                ipfsUrl: answers.ipfs,
            }, { commands, addContract: false });
            if (answers.cleanup) {
                answers.cleanup();
            }
        }
        // Exit with success
        this.exit(0);
    }
}
async function processFromExampleInitForm({ directory: initDirectory, subgraphName: initSubgraphName, }) {
    try {
        const { subgraphName } = await prompt.ask([
            {
                type: 'input',
                name: 'subgraphName',
                message: 'Subgraph slug',
                initial: initSubgraphName,
            },
        ]);
        const { directory } = await prompt.ask([
            {
                type: 'input',
                name: 'directory',
                message: 'Directory to create the subgraph in',
                initial: () => initDirectory || getSubgraphBasename(subgraphName),
            },
        ]);
        return {
            subgraphName,
            directory,
        };
    }
    catch (e) {
        this.error(e, { exit: 1 });
    }
}
async function processInitForm({network: initNetwork, abi: initAbi, abiPath: initAbiPath, directory: initDirectory, source: initContract, indexEvents: initIndexEvents, fromExample: initFromExample, subgraphName: initSubgraphName, contractName: initContractName, startBlock: initStartBlock, spkgPath: initSpkgPath, ipfsUrl, }) {
    try {
        const registry = await loadRegistry();
        const contractService = new ContractService(registry);
        const networks = sortWithPriority(registry.networks, n => n.issuanceRewards, (a, b) => registry.networks.indexOf(a) - registry.networks.indexOf(b));
        const networkToChoice = (n) => ({
            name: n.id,
            value: `${n.id}:${n.shortName}:${n.fullName}`.toLowerCase(),
            hint: `· ${n.id}`,
            message: n.fullName,
        });
        const formatChoices = (choices) => {
            const shown = choices.slice(0, 20);
            const remaining = networks.length - shown.length;
            if (remaining == 0)
                return shown;
            if (shown.length === choices.length) {
                shown.push({
                    name: 'N/A',
                    value: '',
                    hint: '· other network not on the list',
                    message: `Other`,
                });
            }
            return [
                ...shown,
                {
                    name: ``,
                    disabled: true,
                    hint: '',
                    message: `< ${remaining} more - type to filter >`,
                },
            ];
        };
        let network = initNetwork;
        let protocolInstance = new Protocol('ethereum');
        let isComposedSubgraph = false;
        let isSubstreams = false;
        let subgraphName = initSubgraphName ?? '';
        let directory = initDirectory;
        let ipfsNode = '';
        let source = initContract;
        let contractName = initContractName;
        let abiFromFile = undefined;
        let abiFromApi = undefined;
        let startBlock = undefined;
        let spkgPath;
        let spkgCleanup;
        let indexEvents = initIndexEvents;
        const promptManager = new PromptManager();
        promptManager.addStep({
            type: 'autocomplete',
            name: 'networkId',
            required: true,
            message: 'Network',
            inital: initNetwork,
            skip: true,
            choices: formatChoices(networks.map(networkToChoice)),
            format: value => {
                const network = networks.find(n => n.id === value);
                return network
                    ? `${network.fullName}${print.colors.muted(` · ${network.id} · ${network.explorerUrls?.[0] ?? ''}`)}`
                    : value;
            },
            suggest: (input, _) => formatChoices(networks
                .map(networkToChoice)
                .filter(({ value }) => (value ?? '').includes(input.toLowerCase()))),
            validate: value => value === 'N/A' || networks.find(n => n.id === value) ? true : 'Pick a network',
            result: value => {
                initDebugger.extend('processInitForm')('networkId: %O', value);
                const foundNetwork = networks.find(n => n.id === value);
                if (!foundNetwork) {
                    this.log(`
  The chain list is populated from the Networks Registry:

  https://github.com/graphprotocol/networks-registry

  To add a chain to the registry you can create an issue or submit a PR`);
                    process.exit(0);
                }
                network = foundNetwork;
                promptManager.setOptions('protocol', {
                    choices: [
                        {
                            message: 'Smart contract',
                            hint: '· default',
                            name: network.graphNode?.protocol ?? '',
                            value: 'contract',
                        },
                        { message: 'Substreams', name: 'substreams', value: 'substreams' },
                        // { message: 'Subgraph', name: 'subgraph', value: 'subgraph' },
                    ].filter(({ name }) => name),
                });
                return value;
            },
        });
        promptManager.addStep({
            type: 'select',
            name: 'protocol',
            message: 'Source',
            choices: [],
            skip: true,
            validate: name => {
                if (name === 'arweave') {
                    return 'Arweave are only supported via substreams';
                }
                if (name === 'cosmos') {
                    return 'Cosmos chains are only supported via substreams';
                }
                return true;
            },
            format: protocol => {
                switch (protocol) {
                    case '':
                        return '';
                    case 'substreams':
                        return 'Substreams';
                    case 'subgraph':
                        return 'Subgraph';
                    default:
                        return `Smart Contract${print.colors.muted(` · ${protocol}`)}`;
                }
            },
            result: protocol => {
                protocolInstance = new Protocol(protocol);
                isComposedSubgraph = protocolInstance.isComposedSubgraph();
                isSubstreams = protocolInstance.isSubstreams();
                initDebugger.extend('processInitForm')('protocol: %O', protocol);
                return protocol;
            },
        });
        promptManager.addStep({
            type: 'input',
            name: 'subgraphName',
            message: 'Subgraph slug',
            skip: true,
            initial: initSubgraphName,
            validate: value => value.length > 0 || 'Subgraph slug must not be empty',
            result: value => {
                initDebugger.extend('processInitForm')('subgraphName: %O', value);
                subgraphName = value;
                return value;
            },
        });
        promptManager.addStep({
            type: 'input',
            name: 'directory',
            skip: true,
            message: 'Directory to create the subgraph in',
            initial: initDirectory ,
            validate: value => value.length > 0 || 'Directory must not be empty',
            result: value => {
                directory = value;
                initDebugger.extend('processInitForm')('directory: %O', value);
                return value;
            },
        });
        promptManager.addStep({
            type: 'input',
            name: 'source',
            message: () => isComposedSubgraph
                ? 'Source subgraph deployment ID'
                : `Contract ${protocolInstance.getContract()?.identifierName()}`,
            skip: () => initFromExample !== undefined ||
                isSubstreams || initContract !== undefined ||
                (!protocolInstance.hasContract() && !isComposedSubgraph),
            initial: initContract,
            validate: async (value) => {
                if (isComposedSubgraph) {
                    return value.startsWith('Qm') ? true : 'Subgraph deployment ID must start with Qm';
                }
                if (initFromExample !== undefined || !protocolInstance.hasContract()) {
                    return true;
                }
                const { valid, error } = validateContract(value, protocolInstance.getContract());
                return valid ? true : error;
            },
            result: async (address) => {
                initDebugger.extend('processInitForm')("source: '%s'", address);
                if (initFromExample !== undefined ||
                    initAbiPath ||
                    protocolInstance.name !== 'ethereum' // we can only validate against Etherscan API
                ) {
                    source = address;
                    return address;
                }
                // If ABI is not provided, try to fetch it from Etherscan API
                if (protocolInstance.hasABIs() && !initAbi) {
                    abiFromApi = await retryWithPrompt(() => withSpinner('Fetching ABI from contract API...', 'Failed to fetch ABI', 'Warning fetching ABI', () => contractService.getABI(protocolInstance.getABI(), network.id, address)));
                    initDebugger.extend('processInitForm')("abiFromEtherscan len: '%s'", abiFromApi?.name);
                }
                // If startBlock is not provided, try to fetch it from Etherscan API
                if (!initStartBlock) {
                    startBlock = await retryWithPrompt(() => withSpinner('Fetching start block from contract API...', 'Failed to fetch start block', 'Warning fetching start block', () => contractService.getStartBlock(network.id, address)));
                    initDebugger.extend('processInitForm')("startBlockFromEtherscan: '%s'", startBlock);
                }
                // If contract name is not provided, try to fetch it from Etherscan API
                if (!initContractName) {
                    contractName = await retryWithPrompt(() => withSpinner('Fetching contract name from contract API...', 'Failed to fetch contract name', 'Warning fetching contract name', () => contractService.getContractName(network.id, address)));
                    initDebugger.extend('processInitForm')("contractNameFromEtherscan: '%s'", contractName);
                }
                source = address;
                return address;
            },
        });

        startBlock = await retryWithPrompt(() => withSpinner('Fetching start block from contract API...', 'Failed to fetch start block', 'Warning fetching start block', () => contractService.getStartBlock(initNetwork, initContract)));
        initDebugger.extend('processInitForm')("startBlockFromEtherscan: '%s'", startBlock);
        contractName = await retryWithPrompt(() => withSpinner('Fetching contract name from contract API...', 'Failed to fetch contract name', 'Warning fetching contract name', () => contractService.getContractName(initNetwork, initContract)));
        initDebugger.extend('processInitForm')("contractNameFromEtherscan: '%s'", contractName);

        promptManager.addStep({
            type: 'input',
            name: 'ipfs',
            message: `IPFS node to use for fetching subgraph manifest`,
            initial: ipfsUrl,
            skip: () => !isComposedSubgraph,
            result: value => {
                ipfsNode = value;
                initDebugger.extend('processInitForm')('ipfs: %O', value);
                return value;
            },
        });
        promptManager.addStep({
            type: 'input',
            name: 'spkg',
            message: 'Substreams SPKG (local path, IPFS hash, or URL)',
            initial: () => initSpkgPath,
            skip: () => !isSubstreams || !!initSpkgPath,
            validate: async (value) => {
                if (!isSubstreams || !!initSpkgPath)
                    return true;
                return await withSpinner(`Resolving Substreams SPKG file`, `Failed to resolve SPKG file`, `Warnings while resolving SPKG file`, async () => {
                    try {
                        const { path, cleanup } = await resolveFile(value, 'substreams.spkg', 10_000);
                        spkgPath = path;
                        spkgCleanup = cleanup;
                        initDebugger.extend('processInitForm')('spkgPath: %O', path);
                        return true;
                    }
                    catch (e) {
                        return e.message;
                    }
                });
            },
        });
        promptManager.addStep({
            type: 'input',
            name: 'abiFromFile',
            message: 'ABI file (path)',
            initial: initAbiPath,
            skip: () => !protocolInstance.hasABIs() ||
                initFromExample !== undefined ||
                abiFromApi !== undefined ||
                isSubstreams ||
                !!initAbiPath ||
                isComposedSubgraph,
            validate: async (value) => {
                if (initFromExample ||
                    abiFromApi ||
                    !protocolInstance.hasABIs() ||
                    isSubstreams ||
                    isComposedSubgraph) {
                    return true;
                }
                const ABI = protocolInstance.getABI();
                if (initAbiPath)
                    value = initAbiPath;
                try {
                    loadAbiFromFile(ABI, value);
                    return true;
                }
                catch (e) {
                    return e.message;
                }
            },
            result: async (value) => {
                initDebugger.extend('processInitForm')('abiFromFile: %O', value);
                if (initFromExample || abiFromApi || !protocolInstance.hasABIs() || isComposedSubgraph) {
                    return null;
                }
                const ABI = protocolInstance.getABI();
                if (initAbiPath)
                    value = initAbiPath;
                try {
                    abiFromFile = loadAbiFromFile(ABI, value);
                    return value;
                }
                catch (e) {
                    return e.message;
                }
            },
        });
        promptManager.addStep({
            type: 'input',
            name: 'startBlock',
            message: 'Start block',
            initial: () => initStartBlock || startBlock || '0',
            skip: () => initFromExample !== undefined || isSubstreams || startBlock !== '0',
            // validate: value => initFromExample !== undefined ||
            //     isSubstreams ||
            //     parseInt(value) >= 0 ||
            //     'Invalid start block',
            // result: value => {
            //     startBlock = value;
            //     initDebugger.extend('processInitForm')('startBlock: %O', value);
            //     return value;
            // },
        });
        promptManager.addStep({
            type: 'input',
            name: 'contractName',
            message: 'Contract name',
            initial: () => initContractName || contractName || 'Contract',
            skip: () => initFromExample !== undefined || !protocolInstance.hasContract() || isSubstreams || true,
            // validate: value => initFromExample !== undefined ||
            //     !protocolInstance.hasContract() ||
            //     isSubstreams ||
            //     value.length > 0 ||
            //     'Contract name must not be empty',
            // result: value => {
            //     contractName = value;
            //     initDebugger.extend('processInitForm')('contractName: %O', value);
            //     return value;
            // },
        });
        promptManager.addStep({
            type: 'confirm',
            name: 'indexEvents',
            message: 'Index contract events as entities',
            initial: true,
            skip: () => !!initIndexEvents || isSubstreams || isComposedSubgraph || true,
            result: value => {
                indexEvents = String(value) === 'true';
                initDebugger.extend('processInitForm')('indexEvents: %O', indexEvents);
                return value;
            },
        });
        await promptManager.executeInteractive();
        return {
            abi: (abiFromApi || abiFromFile),
            protocolInstance,
            subgraphName,
            directory: directory,
            startBlock: startBlock,
            fromExample: !!initFromExample,
            network: network.id,
            contractName: contractName,
            source: source,
            indexEvents,
            ipfs: ipfsNode,
            spkgPath,
            cleanup: spkgCleanup,
        };
    }
    catch (e) {
        this.error(e, { exit: 1 });
    }
}
const loadAbiFromFile = (ABI, filename) => {
    const exists = filesystem.exists(filename);
    if (!exists) {
        throw Error('File does not exist.');
    }
    else if (exists === 'dir') {
        throw Error('Path points to a directory, not a file.');
    }
    else if (exists === 'other') {
        throw Error('Not sure what this path points to.');
    }
    else {
        return ABI.load('Contract', filename);
    }
};
// Inspired from: https://github.com/graphprotocol/graph-tooling/issues/1450#issuecomment-1713992618
async function isInRepo() {
    try {
        const result = await system.run('git rev-parse --is-inside-work-tree');
        // It seems like we are returning "true\n" instead of "true".
        // Don't think it is great idea to check for new line character here.
        // So best to just check if the result includes "true".
        return result.includes('true');
    }
    catch (err) {
        if (err.stderr.includes('not a git repository')) {
            return false;
        }
        throw Error(err.stderr);
    }
}
const initRepository = async (directory) => await withSpinner(`Initialize subgraph repository`, `Failed to initialize subgraph repository`, `Warnings while initializing subgraph repository`, async () => {
    // Remove .git dir in --from-example mode; in --from-contract, we're
    // starting from an empty directory
    const gitDir = path.join(directory, '.git');
    if (filesystem.exists(gitDir)) {
        filesystem.remove(gitDir);
    }
    if (await isInRepo()) {
        await system.run('git add --all', { cwd: directory });
        await system.run('git commit -m "Initialize subgraph"', {
            cwd: directory,
        });
    }
    else {
        await system.run('git init', { cwd: directory });
        await system.run('git add --all', { cwd: directory });
        await system.run('git commit -m "Initial commit"', {
            cwd: directory,
        });
    }
    return true;
});
const installDependencies = async (directory, commands) => await withSpinner(`Install dependencies with ${commands.install}`, `Failed to install dependencies`, `Warnings while installing dependencies`, async () => {
    if (process.env.GRAPH_CLI_TESTS) {
        await system.run(commands.link, { cwd: directory });
    }
    await system.run(commands.install, { cwd: directory });
    return true;
});
const runCodegen = async (directory, codegenCommand) => await withSpinner(`Generate ABI and schema types with ${codegenCommand}`, `Failed to generate code from ABI and GraphQL schema`, `Warnings while generating code from ABI and GraphQL schema`, async () => {
    await system.run(codegenCommand, { cwd: directory });
    return true;
});
function printNextSteps({ subgraphName, directory }, { commands, }) {
    const relativeDir = path.relative(process.cwd(), directory);
    // Print instructions
    this.log(`
Subgraph ${subgraphName} created in ${relativeDir}
`);
    this.log(`Next steps:

  1. Run \`graph auth\` to authenticate with your deploy key.

  2. Type \`cd ${relativeDir}\` to enter the subgraph.

  3. Run \`${commands.deploy}\` to deploy the subgraph.

Make sure to visit the documentation on https://thegraph.com/docs/ for further information.`);
}
async function initSubgraphFromExample({ fromExample, subgraphName, directory, skipInstall, skipGit, }, { commands, }) {
    if (filesystem.exists(directory)) {
        const overwrite = await prompt
            .confirm('Directory already exists, do you want to initialize the subgraph here (files will be overwritten) ?', false)
            .catch(() => false);
        if (!overwrite) {
            this.exit(1);
        }
    }
    // Clone the example subgraph repository
    const cloned = await withSpinner(`Cloning example subgraph`, `Failed to clone example subgraph`, `Warnings while cloning example subgraph`, async () => {
        // Create a temporary directory
        const prefix = path.join(os.tmpdir(), 'example-subgraph-');
        const tmpDir = fs.mkdtempSync(prefix);
        try {
            await system.run(`git clone https://github.com/graphprotocol/graph-tooling ${tmpDir}`);
            // If an example is not specified, use the default one
            if (fromExample === undefined || fromExample === true) {
                fromExample = DEFAULT_EXAMPLE_SUBGRAPH;
            }
            // Legacy purposes when everything existed in examples repo
            if (fromExample === 'ethereum/gravatar') {
                fromExample = DEFAULT_EXAMPLE_SUBGRAPH;
            }
            const exampleSubgraphPath = path.join(tmpDir, 'examples', String(fromExample));
            if (!filesystem.exists(exampleSubgraphPath)) {
                return { result: false, error: `Example not found: ${fromExample}` };
            }
            filesystem.copy(exampleSubgraphPath, directory, { overwrite: true });
            return true;
        }
        finally {
            filesystem.remove(tmpDir);
        }
    });
    if (!cloned) {
        this.exit(1);
    }
    const networkConf = await initNetworksConfig(directory, 'address');
    if (networkConf !== true) {
        this.exit(1);
    }
    // Update package.json to match the subgraph name
    const prepared = await withSpinner(`Update subgraph name and commands in package.json`, `Failed to update subgraph name and commands in package.json`, `Warnings while updating subgraph name and commands in package.json`, async () => {
        try {
            // Load package.json
            const pkgJsonFilename = filesystem.path(directory, 'package.json');
            const pkgJson = await filesystem.read(pkgJsonFilename, 'json');
            pkgJson.name = getSubgraphBasename(subgraphName);
            for (const name of Object.keys(pkgJson.scripts)) {
                pkgJson.scripts[name] = pkgJson.scripts[name].replace('example', subgraphName);
            }
            delete pkgJson['license'];
            delete pkgJson['repository'];
            // Remove example's cli in favor of the local one (added via `npm link`)
            if (process.env.GRAPH_CLI_TESTS) {
                delete pkgJson['devDependencies']['@graphprotocol/graph-cli'];
            }
            // Write package.json
            filesystem.write(pkgJsonFilename, pkgJson, { jsonIndent: 2 });
            return true;
        }
        catch (e) {
            filesystem.remove(directory);
            this.error(`Failed to preconfigure the subgraph: ${e}`);
        }
    });
    if (!prepared) {
        this.exit(1);
    }
    // Initialize a fresh Git repository
    if (!skipGit) {
        const repo = await initRepository(directory);
        if (repo !== true) {
            this.exit(1);
        }
    }
    // Install dependencies
    if (!skipInstall) {
        const installed = await installDependencies(directory, commands);
        if (installed !== true) {
            this.exit(1);
        }
    }
    // Run code-generation
    const codegen = await runCodegen(directory, commands.codegen);
    if (codegen !== true) {
        this.exit(1);
    }
    printNextSteps.bind(this)({ subgraphName, directory }, { commands });
}
async function initSubgraphFromContract({ protocolInstance, subgraphName, directory, abi, network, source, indexEvents, contractName, node, startBlock, spkgPath, skipInstall, skipGit, ipfsUrl, }, { commands, addContract, }) {
    const isComposedSubgraph = protocolInstance.isComposedSubgraph();
    if (filesystem.exists(directory)) {
        const overwrite = await prompt
            .confirm('Directory already exists, do you want to initialize the subgraph here (files will be overwritten) ?', false)
            .catch(() => false);
        if (!overwrite) {
            this.exit(1);
        }
    }
    let entities;
    if (isComposedSubgraph) {
        try {
            const ipfsClient = createIpfsClient({
                url: appendApiVersionForGraph(ipfsUrl),
                headers: {
                    ...GRAPH_CLI_SHARED_HEADERS,
                },
            });
            const schemaString = await loadSubgraphSchemaFromIPFS(ipfsClient, source);
            const schema = await Schema.loadFromString(schemaString);
            entities = schema.getEntityNames();
        }
        catch (e) {
            this.error(`Failed to load and parse subgraph schema: ${e.message}`, { exit: 1 });
        }
    }
    if (!protocolInstance.isComposedSubgraph() &&
        protocolInstance.hasABIs() &&
        (abiEvents(abi).size === 0 ||
            // @ts-expect-error TODO: the abiEvents result is expected to be a List, how's it an array?
            abiEvents(abi).length === 0)) {
        // Fail if the ABI does not contain any events
        this.error(`ABI does not contain any events`, { exit: 1 });
    }
    // Scaffold subgraph
    const scaffold = await withSpinner(`Create subgraph scaffold`, `Failed to create subgraph scaffold`, `Warnings while creating subgraph scaffold`, async (spinner) => {
        const scaffold = await generateScaffold({
            protocolInstance,
            subgraphName,
            abi,
            network,
            source,
            indexEvents,
            contractName,
            startBlock,
            node,
            spkgPath,
            entities,
        }, spinner);
        await writeScaffold(scaffold, directory, spinner);
        return true;
    });
    if (scaffold !== true) {
        this.exit(1);
    }
    if (protocolInstance.hasContract()) {
        const identifierName = protocolInstance.getContract().identifierName();
        const networkConf = await initNetworksConfig(directory, identifierName);
        if (networkConf !== true) {
            this.exit(1);
        }
    }
    // Initialize a fresh Git repository
    if (!skipGit) {
        const repo = await initRepository(directory);
        if (repo !== true) {
            this.exit(1);
        }
    }
    if (!skipInstall) {
        // Install dependencies
        const installed = await installDependencies(directory, commands);
        if (installed !== true) {
            this.exit(1);
        }
    }
    // Substreams we have nothing to install or generate
    if (!protocolInstance.isSubstreams()) {
        // Run code-generation
        const codegen = await runCodegen(directory, commands.codegen);
        if (codegen !== true) {
            this.exit(1);
        }
        while (addContract) {
            addContract = await addAnotherContract
                .bind(this)({
                    protocolInstance,
                    directory,
                })
                .catch(() => false);
        }
    }
    // printNextSteps.bind(this)({ subgraphName, directory }, { commands });
}
async function addAnotherContract({ protocolInstance, directory, }) {
    const { addAnother } = await prompt.ask([
        {
            type: 'confirm',
            name: 'addAnother',
            message: () => 'Add another contract?',
            initial: false,
            required: true,
        },
    ]);
    if (!addAnother)
        return false;
    const ProtocolContract = protocolInstance.getContract();
    const { contract } = await prompt.ask([
        {
            type: 'input',
            name: 'contract',
            initial: ProtocolContract.identifierName(),
            required: true,
            message: () => `\nContract ${ProtocolContract.identifierName()}`,
            validate: value => {
                const { valid, error } = validateContract(value, ProtocolContract);
                return valid ? true : error;
            },
        },
    ]);
    const cwd = process.cwd();
    try {
        if (fs.existsSync(directory)) {
            process.chdir(directory);
        }
        await AddCommand.run([contract]);
    }
    catch (e) {
        this.error(e);
    }
    process.chdir(cwd);
    return true;
}
