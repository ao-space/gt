declare const _default: import("vue").DefineComponent<{
    readonly type: import("../../../utils").EpPropFinalized<StringConstructor, "" | "success" | "warning" | "info" | "primary" | "danger", unknown, "", boolean>;
    readonly size: import("../../../utils").EpPropFinalized<StringConstructor, "" | "default" | "small" | "large", unknown, "", boolean>;
    readonly truncated: {
        readonly type: import("vue").PropType<import("../../../utils").EpPropMergeType<BooleanConstructor, unknown, unknown>>;
        readonly required: false;
        readonly validator: ((val: unknown) => boolean) | undefined;
        __epPropKey: true;
    };
    readonly tag: import("../../../utils").EpPropFinalized<StringConstructor, unknown, unknown, "span", boolean>;
}, {
    props: Readonly<import("@vue/shared").LooseRequired<Readonly<import("vue").ExtractPropTypes<{
        readonly type: import("../../../utils").EpPropFinalized<StringConstructor, "" | "success" | "warning" | "info" | "primary" | "danger", unknown, "", boolean>;
        readonly size: import("../../../utils").EpPropFinalized<StringConstructor, "" | "default" | "small" | "large", unknown, "", boolean>;
        readonly truncated: {
            readonly type: import("vue").PropType<import("../../../utils").EpPropMergeType<BooleanConstructor, unknown, unknown>>;
            readonly required: false;
            readonly validator: ((val: unknown) => boolean) | undefined;
            __epPropKey: true;
        };
        readonly tag: import("../../../utils").EpPropFinalized<StringConstructor, unknown, unknown, "span", boolean>;
    }>> & {
        [x: string & `on${string}`]: ((...args: any[]) => any) | ((...args: unknown[]) => any) | undefined;
    }>>;
    textSize: import("vue").ComputedRef<"" | "default" | "small" | "large">;
    ns: {
        namespace: import("vue").ComputedRef<string>;
        b: (blockSuffix?: string) => string;
        e: (element?: string | undefined) => string;
        m: (modifier?: string | undefined) => string;
        be: (blockSuffix?: string | undefined, element?: string | undefined) => string;
        em: (element?: string | undefined, modifier?: string | undefined) => string;
        bm: (blockSuffix?: string | undefined, modifier?: string | undefined) => string;
        bem: (blockSuffix?: string | undefined, element?: string | undefined, modifier?: string | undefined) => string;
        is: {
            (name: string, state: boolean | undefined): string;
            (name: string): string;
        };
        cssVar: (object: Record<string, string>) => Record<string, string>;
        cssVarName: (name: string) => string;
        cssVarBlock: (object: Record<string, string>) => Record<string, string>;
        cssVarBlockName: (name: string) => string;
    };
    textKls: import("vue").ComputedRef<string[]>;
}, unknown, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, Record<string, any>, string, import("vue").VNodeProps & import("vue").AllowedComponentProps & import("vue").ComponentCustomProps, Readonly<import("vue").ExtractPropTypes<{
    readonly type: import("../../../utils").EpPropFinalized<StringConstructor, "" | "success" | "warning" | "info" | "primary" | "danger", unknown, "", boolean>;
    readonly size: import("../../../utils").EpPropFinalized<StringConstructor, "" | "default" | "small" | "large", unknown, "", boolean>;
    readonly truncated: {
        readonly type: import("vue").PropType<import("../../../utils").EpPropMergeType<BooleanConstructor, unknown, unknown>>;
        readonly required: false;
        readonly validator: ((val: unknown) => boolean) | undefined;
        __epPropKey: true;
    };
    readonly tag: import("../../../utils").EpPropFinalized<StringConstructor, unknown, unknown, "span", boolean>;
}>>, {
    readonly type: import("../../../utils").EpPropMergeType<StringConstructor, "" | "success" | "warning" | "info" | "primary" | "danger", unknown>;
    readonly size: import("../../../utils").EpPropMergeType<StringConstructor, "" | "default" | "small" | "large", unknown>;
    readonly tag: string;
}>;
export default _default;
