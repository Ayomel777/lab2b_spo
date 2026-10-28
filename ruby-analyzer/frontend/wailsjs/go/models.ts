export namespace analyzer {
	
	export class VertexComplexity {
	    vertexId: number;
	    complexity: number;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new VertexComplexity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vertexId = source["vertexId"];
	        this.complexity = source["complexity"];
	        this.kind = source["kind"];
	    }
	}
	export class SubgraphInfo {
	    selectVertex: number;
	    transitions: number[];
	    selectCost: number;
	    subgraphVerts: number[];
	    lowerBound: number;
	
	    static createFrom(source: any = {}) {
	        return new SubgraphInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selectVertex = source["selectVertex"];
	        this.transitions = source["transitions"];
	        this.selectCost = source["selectCost"];
	        this.subgraphVerts = source["subgraphVerts"];
	        this.lowerBound = source["lowerBound"];
	    }
	}
	export class BoundaryResult {
	    sa: number;
	    so: number;
	    subs: SubgraphInfo[];
	    vertices: VertexComplexity[];
	
	    static createFrom(source: any = {}) {
	        return new BoundaryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sa = source["sa"];
	        this.so = source["so"];
	        this.subs = this.convertValues(source["subs"], SubgraphInfo);
	        this.vertices = this.convertValues(source["vertices"], VertexComplexity);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CFGNode {
	    id: number;
	    kind: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new CFGNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.label = source["label"];
	    }
	}
	export class CFG {
	    nodes: CFGNode[];
	
	    static createFrom(source: any = {}) {
	        return new CFG(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodes = this.convertValues(source["nodes"], CFGNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Result {
	    boundary: BoundaryResult;
	    cfg?: CFG;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.boundary = this.convertValues(source["boundary"], BoundaryResult);
	        this.cfg = this.convertValues(source["cfg"], CFG);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

