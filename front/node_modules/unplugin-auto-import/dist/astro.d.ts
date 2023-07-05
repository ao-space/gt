import { Options } from './types.js';
import '@antfu/utils';
import '@rollup/pluginutils';
import 'unimport';
import 'unimport/dist/types-2cb9c8f5';

declare function export_default(options: Options): {
    name: string;
    hooks: {
        'astro:config:setup': (astro: any) => Promise<void>;
    };
};

export { export_default as default };
