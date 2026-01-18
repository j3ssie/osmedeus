(globalThis.TURBOPACK||(globalThis.TURBOPACK=[])).push(["object"==typeof document?document.currentScript:void 0,57763,e=>{"use strict";var t=function(e,t){var n,o="";for(n=0;n<t;n+=1)o+=e;return o},n=function(e){return 0===e&&-1/0==1/e};function o(e,t){var n="",o=e.reason||"(unknown reason)";return e.mark?(e.mark.name&&(n+='in "'+e.mark.name+'" '),n+="("+(e.mark.line+1)+":"+(e.mark.column+1)+")",!t&&e.mark.snippet&&(n+="\n\n"+e.mark.snippet),o+" "+n):o}function r(e,t){Error.call(this),this.name="YAMLException",this.reason=e,this.mark=t,this.message=o(this,!1),Error.captureStackTrace?Error.captureStackTrace(this,this.constructor):this.stack=Error().stack||""}function i(e,t,n,o,r){var i="",a="",s=Math.floor(r/2)-1;return o-t>s&&(t=o-s+(i=" ... ").length),n-o>s&&(n=o+s-(a=" ...").length),{str:i+e.slice(t,n).replace(/\t/g,"→")+a,pos:o-t+i.length}}function a(e,n){return t(" ",n-e.length)+e}r.prototype=Object.create(Error.prototype),r.prototype.constructor=r,r.prototype.toString=function(e){return this.name+": "+o(this,e)};var s=function(e,n){if(n=Object.create(n||null),!e.buffer)return null;n.maxLength||(n.maxLength=79),"number"!=typeof n.indent&&(n.indent=1),"number"!=typeof n.linesBefore&&(n.linesBefore=3),"number"!=typeof n.linesAfter&&(n.linesAfter=2);for(var o=/\r?\n|\r|\0/g,r=[0],s=[],l=-1;u=o.exec(e.buffer);)s.push(u.index),r.push(u.index+u[0].length),e.position<=u.index&&l<0&&(l=r.length-2);l<0&&(l=r.length-1);var u,c,p,d="",m=Math.min(e.line+n.linesAfter,s.length).toString().length,f=n.maxLength-(n.indent+m+3);for(c=1;c<=n.linesBefore&&!(l-c<0);c++)p=i(e.buffer,r[l-c],s[l-c],e.position-(r[l]-r[l-c]),f),d=t(" ",n.indent)+a((e.line-c+1).toString(),m)+" | "+p.str+"\n"+d;for(p=i(e.buffer,r[l],s[l],e.position,f),d+=t(" ",n.indent)+a((e.line+1).toString(),m)+" | "+p.str+"\n"+t("-",n.indent+m+3+p.pos)+"^\n",c=1;c<=n.linesAfter&&!(l+c>=s.length);c++)p=i(e.buffer,r[l+c],s[l+c],e.position-(r[l]-r[l+c]),f),d+=t(" ",n.indent)+a((e.line+c+1).toString(),m)+" | "+p.str+"\n";return d.replace(/\n$/,"")},l=["kind","multi","resolve","construct","instanceOf","predicate","represent","representName","defaultStyle","styleAliases"],u=["scalar","sequence","mapping"],c=function(e,t){var n,o;if(Object.keys(t=t||{}).forEach(function(t){if(-1===l.indexOf(t))throw new r('Unknown option "'+t+'" is met in definition of "'+e+'" YAML type.')}),this.options=t,this.tag=e,this.kind=t.kind||null,this.resolve=t.resolve||function(){return!0},this.construct=t.construct||function(e){return e},this.instanceOf=t.instanceOf||null,this.predicate=t.predicate||null,this.represent=t.represent||null,this.representName=t.representName||null,this.defaultStyle=t.defaultStyle||null,this.multi=t.multi||!1,this.styleAliases=(n=t.styleAliases||null,o={},null!==n&&Object.keys(n).forEach(function(e){n[e].forEach(function(t){o[String(t)]=e})}),o),-1===u.indexOf(this.kind))throw new r('Unknown kind "'+this.kind+'" is specified for "'+e+'" YAML type.')};function p(e,t){var n=[];return e[t].forEach(function(e){var t=n.length;n.forEach(function(n,o){n.tag===e.tag&&n.kind===e.kind&&n.multi===e.multi&&(t=o)}),n[t]=e}),n}function d(e){return this.extend(e)}d.prototype.extend=function(e){var t=[],n=[];if(e instanceof c)n.push(e);else if(Array.isArray(e))n=n.concat(e);else if(e&&(Array.isArray(e.implicit)||Array.isArray(e.explicit)))e.implicit&&(t=t.concat(e.implicit)),e.explicit&&(n=n.concat(e.explicit));else throw new r("Schema.extend argument should be a Type, [ Type ], or a schema definition ({ implicit: [...], explicit: [...] })");t.forEach(function(e){if(!(e instanceof c))throw new r("Specified list of YAML types (or a single Type object) contains a non-Type object.");if(e.loadKind&&"scalar"!==e.loadKind)throw new r("There is a non-scalar type in the implicit list of a schema. Implicit resolving of such types is not supported.");if(e.multi)throw new r("There is a multi type in the implicit list of a schema. Multi tags can only be listed as explicit.")}),n.forEach(function(e){if(!(e instanceof c))throw new r("Specified list of YAML types (or a single Type object) contains a non-Type object.")});var o=Object.create(d.prototype);return o.implicit=(this.implicit||[]).concat(t),o.explicit=(this.explicit||[]).concat(n),o.compiledImplicit=p(o,"implicit"),o.compiledExplicit=p(o,"explicit"),o.compiledTypeMap=function(){var e,t,n={scalar:{},sequence:{},mapping:{},fallback:{},multi:{scalar:[],sequence:[],mapping:[],fallback:[]}};function o(e){e.multi?(n.multi[e.kind].push(e),n.multi.fallback.push(e)):n[e.kind][e.tag]=n.fallback[e.tag]=e}for(e=0,t=arguments.length;e<t;e+=1)arguments[e].forEach(o);return n}(o.compiledImplicit,o.compiledExplicit),o};var m=new c("tag:yaml.org,2002:str",{kind:"scalar",construct:function(e){return null!==e?e:""}}),f=new c("tag:yaml.org,2002:seq",{kind:"sequence",construct:function(e){return null!==e?e:[]}}),h=new c("tag:yaml.org,2002:map",{kind:"mapping",construct:function(e){return null!==e?e:{}}}),g=new d({explicit:[m,f,h]}),y=new c("tag:yaml.org,2002:null",{kind:"scalar",resolve:function(e){if(null===e)return!0;var t=e.length;return 1===t&&"~"===e||4===t&&("null"===e||"Null"===e||"NULL"===e)},construct:function(){return null},predicate:function(e){return null===e},represent:{canonical:function(){return"~"},lowercase:function(){return"null"},uppercase:function(){return"NULL"},camelcase:function(){return"Null"},empty:function(){return""}},defaultStyle:"lowercase"}),w=new c("tag:yaml.org,2002:bool",{kind:"scalar",resolve:function(e){if(null===e)return!1;var t=e.length;return 4===t&&("true"===e||"True"===e||"TRUE"===e)||5===t&&("false"===e||"False"===e||"FALSE"===e)},construct:function(e){return"true"===e||"True"===e||"TRUE"===e},predicate:function(e){return"[object Boolean]"===Object.prototype.toString.call(e)},represent:{lowercase:function(e){return e?"true":"false"},uppercase:function(e){return e?"TRUE":"FALSE"},camelcase:function(e){return e?"True":"False"}},defaultStyle:"lowercase"}),b=new c("tag:yaml.org,2002:int",{kind:"scalar",resolve:function(e){if(null===e)return!1;var t,n,o,r,i=e.length,a=0,s=!1;if(!i)return!1;if(("-"===(r=e[a])||"+"===r)&&(r=e[++a]),"0"===r){if(a+1===i)return!0;if("b"===(r=e[++a])){for(a++;a<i;a++)if("_"!==(r=e[a])){if("0"!==r&&"1"!==r)return!1;s=!0}return s&&"_"!==r}if("x"===r){for(a++;a<i;a++)if("_"!==(r=e[a])){if(!(48<=(t=e.charCodeAt(a))&&t<=57||65<=t&&t<=70||97<=t&&t<=102))return!1;s=!0}return s&&"_"!==r}if("o"===r){for(a++;a<i;a++)if("_"!==(r=e[a])){if(!(48<=(n=e.charCodeAt(a))&&n<=55))return!1;s=!0}return s&&"_"!==r}}if("_"===r)return!1;for(;a<i;a++)if("_"!==(r=e[a])){if(!(48<=(o=e.charCodeAt(a))&&o<=57))return!1;s=!0}return!!s&&"_"!==r},construct:function(e){var t,n=e,o=1;if(-1!==n.indexOf("_")&&(n=n.replace(/_/g,"")),("-"===(t=n[0])||"+"===t)&&("-"===t&&(o=-1),t=(n=n.slice(1))[0]),"0"===n)return 0;if("0"===t){if("b"===n[1])return o*parseInt(n.slice(2),2);if("x"===n[1])return o*parseInt(n.slice(2),16);if("o"===n[1])return o*parseInt(n.slice(2),8)}return o*parseInt(n,10)},predicate:function(e){return"[object Number]"===Object.prototype.toString.call(e)&&e%1==0&&!n(e)},represent:{binary:function(e){return e>=0?"0b"+e.toString(2):"-0b"+e.toString(2).slice(1)},octal:function(e){return e>=0?"0o"+e.toString(8):"-0o"+e.toString(8).slice(1)},decimal:function(e){return e.toString(10)},hexadecimal:function(e){return e>=0?"0x"+e.toString(16).toUpperCase():"-0x"+e.toString(16).toUpperCase().slice(1)}},defaultStyle:"decimal",styleAliases:{binary:[2,"bin"],octal:[8,"oct"],decimal:[10,"dec"],hexadecimal:[16,"hex"]}}),_=RegExp("^(?:[-+]?(?:[0-9][0-9_]*)(?:\\.[0-9_]*)?(?:[eE][-+]?[0-9]+)?|\\.[0-9_]+(?:[eE][-+]?[0-9]+)?|[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"),v=/^[-+]?[0-9]+e/,k=new c("tag:yaml.org,2002:float",{kind:"scalar",resolve:function(e){return null!==e&&!!_.test(e)&&"_"!==e[e.length-1]},construct:function(e){var t,n;return(n="-"===(t=e.replace(/_/g,"").toLowerCase())[0]?-1:1,"+-".indexOf(t[0])>=0&&(t=t.slice(1)),".inf"===t)?1===n?1/0:-1/0:".nan"===t?NaN:n*parseFloat(t,10)},predicate:function(e){return"[object Number]"===Object.prototype.toString.call(e)&&(e%1!=0||n(e))},represent:function(e,t){var o;if(isNaN(e))switch(t){case"lowercase":return".nan";case"uppercase":return".NAN";case"camelcase":return".NaN"}else if(1/0===e)switch(t){case"lowercase":return".inf";case"uppercase":return".INF";case"camelcase":return".Inf"}else if(-1/0===e)switch(t){case"lowercase":return"-.inf";case"uppercase":return"-.INF";case"camelcase":return"-.Inf"}else if(n(e))return"-0.0";return o=e.toString(10),v.test(o)?o.replace("e",".e"):o},defaultStyle:"lowercase"}),x=g.extend({implicit:[y,w,b,k]}),S=RegExp("^([0-9][0-9][0-9][0-9])-([0-9][0-9])-([0-9][0-9])$"),A=RegExp("^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:[Tt]|[ \\t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\\.([0-9]*))?(?:[ \\t]*(Z|([-+])([0-9][0-9]?)(?::([0-9][0-9]))?))?$"),O=new c("tag:yaml.org,2002:timestamp",{kind:"scalar",resolve:function(e){return null!==e&&(null!==S.exec(e)||null!==A.exec(e))},construct:function(e){var t,n,o,r,i,a,s,l,u=0,c=null;if(null===(t=S.exec(e))&&(t=A.exec(e)),null===t)throw Error("Date resolve error");if(n=+t[1],o=t[2]-1,r=+t[3],!t[4])return new Date(Date.UTC(n,o,r));if(i=+t[4],a=+t[5],s=+t[6],t[7]){for(u=t[7].slice(0,3);u.length<3;)u+="0";u*=1}return t[9]&&(c=(60*t[10]+ +(t[11]||0))*6e4,"-"===t[9]&&(c=-c)),l=new Date(Date.UTC(n,o,r,i,a,s,u)),c&&l.setTime(l.getTime()-c),l},instanceOf:Date,represent:function(e){return e.toISOString()}}),C=new c("tag:yaml.org,2002:merge",{kind:"scalar",resolve:function(e){return"<<"===e||null===e}}),T="ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=\n\r",E=new c("tag:yaml.org,2002:binary",{kind:"scalar",resolve:function(e){if(null===e)return!1;var t,n,o=0,r=e.length;for(n=0;n<r;n++)if(!((t=T.indexOf(e.charAt(n)))>64)){if(t<0)return!1;o+=6}return o%8==0},construct:function(e){var t,n,o=e.replace(/[\r\n=]/g,""),r=o.length,i=0,a=[];for(t=0;t<r;t++)t%4==0&&t&&(a.push(i>>16&255),a.push(i>>8&255),a.push(255&i)),i=i<<6|T.indexOf(o.charAt(t));return 0==(n=r%4*6)?(a.push(i>>16&255),a.push(i>>8&255),a.push(255&i)):18===n?(a.push(i>>10&255),a.push(i>>2&255)):12===n&&a.push(i>>4&255),new Uint8Array(a)},predicate:function(e){return"[object Uint8Array]"===Object.prototype.toString.call(e)},represent:function(e){var t,n,o="",r=0,i=e.length;for(t=0;t<i;t++)t%3==0&&t&&(o+=T[r>>18&63],o+=T[r>>12&63],o+=T[r>>6&63],o+=T[63&r]),r=(r<<8)+e[t];return 0==(n=i%3)?(o+=T[r>>18&63],o+=T[r>>12&63],o+=T[r>>6&63],o+=T[63&r]):2===n?(o+=T[r>>10&63],o+=T[r>>4&63],o+=T[r<<2&63],o+=T[64]):1===n&&(o+=T[r>>2&63],o+=T[r<<4&63],o+=T[64],o+=T[64]),o}}),I=Object.prototype.hasOwnProperty,F=Object.prototype.toString,M=new c("tag:yaml.org,2002:omap",{kind:"sequence",resolve:function(e){if(null===e)return!0;var t,n,o,r,i,a=[];for(t=0,n=e.length;t<n;t+=1){if(o=e[t],i=!1,"[object Object]"!==F.call(o))return!1;for(r in o)if(I.call(o,r))if(i)return!1;else i=!0;if(!i||-1!==a.indexOf(r))return!1;a.push(r)}return!0},construct:function(e){return null!==e?e:[]}}),L=Object.prototype.toString,R=new c("tag:yaml.org,2002:pairs",{kind:"sequence",resolve:function(e){var t,n,o,r,i;if(null===e)return!0;for(t=0,i=Array(e.length),n=e.length;t<n;t+=1){if(o=e[t],"[object Object]"!==L.call(o)||1!==(r=Object.keys(o)).length)return!1;i[t]=[r[0],o[r[0]]]}return!0},construct:function(e){var t,n,o,r,i;if(null===e)return[];for(t=0,i=Array(e.length),n=e.length;t<n;t+=1)r=Object.keys(o=e[t]),i[t]=[r[0],o[r[0]]];return i}}),P=Object.prototype.hasOwnProperty,j=new c("tag:yaml.org,2002:set",{kind:"mapping",resolve:function(e){var t;if(null===e)return!0;for(t in e)if(P.call(e,t)&&null!==e[t])return!1;return!0},construct:function(e){return null!==e?e:{}}}),N=x.extend({implicit:[O,C],explicit:[E,M,R,j]}),D=Object.prototype.hasOwnProperty,q=/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x84\x86-\x9F\uFFFE\uFFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?:[^\uD800-\uDBFF]|^)[\uDC00-\uDFFF]/,H=/[\x85\u2028\u2029]/,U=/[,\[\]\{\}]/,W=/^(?:!|!!|![a-z\-]+!)$/i,Y=/^(?:!|[^,\[\]\{\}])(?:%[0-9a-f]{2}|[0-9a-z\-#;\/\?:@&=\+\$,_\.!~\*'\(\)\[\]])*$/i;function G(e){return Object.prototype.toString.call(e)}function B(e){return 10===e||13===e}function $(e){return 9===e||32===e}function K(e){return 9===e||32===e||10===e||13===e}function J(e){return 44===e||91===e||93===e||123===e||125===e}function V(e){return 48===e?"\0":97===e?"\x07":98===e?"\b":116===e||9===e?"	":110===e?"\n":118===e?"\v":102===e?"\f":114===e?"\r":101===e?"\x1b":32===e?" ":34===e?'"':47===e?"/":92===e?"\\":78===e?"":95===e?" ":76===e?"\u2028":80===e?"\u2029":""}function z(e,t,n){"__proto__"===t?Object.defineProperty(e,t,{configurable:!0,enumerable:!0,writable:!0,value:n}):e[t]=n}for(var X=Array(256),Q=Array(256),Z=0;Z<256;Z++)X[Z]=+!!V(Z),Q[Z]=V(Z);function ee(e,t){this.input=e,this.filename=t.filename||null,this.schema=t.schema||N,this.onWarning=t.onWarning||null,this.legacy=t.legacy||!1,this.json=t.json||!1,this.listener=t.listener||null,this.implicitTypes=this.schema.compiledImplicit,this.typeMap=this.schema.compiledTypeMap,this.length=e.length,this.position=0,this.line=0,this.lineStart=0,this.lineIndent=0,this.firstTabInLine=-1,this.documents=[]}function et(e,t){var n={name:e.filename,buffer:e.input.slice(0,-1),position:e.position,line:e.line,column:e.position-e.lineStart};return n.snippet=s(n),new r(t,n)}function en(e,t){throw et(e,t)}function eo(e,t){e.onWarning&&e.onWarning.call(null,et(e,t))}var er={YAML:function(e,t,n){var o,r,i;null!==e.version&&en(e,"duplication of %YAML directive"),1!==n.length&&en(e,"YAML directive accepts exactly one argument"),null===(o=/^([0-9]+)\.([0-9]+)$/.exec(n[0]))&&en(e,"ill-formed argument of the YAML directive"),r=parseInt(o[1],10),i=parseInt(o[2],10),1!==r&&en(e,"unacceptable YAML version of the document"),e.version=n[0],e.checkLineBreaks=i<2,1!==i&&2!==i&&eo(e,"unsupported YAML version of the document")},TAG:function(e,t,n){var o,r;2!==n.length&&en(e,"TAG directive accepts exactly two arguments"),o=n[0],r=n[1],W.test(o)||en(e,"ill-formed tag handle (first argument) of the TAG directive"),D.call(e.tagMap,o)&&en(e,'there is a previously declared suffix for "'+o+'" tag handle'),Y.test(r)||en(e,"ill-formed tag prefix (second argument) of the TAG directive");try{r=decodeURIComponent(r)}catch(t){en(e,"tag prefix is malformed: "+r)}e.tagMap[o]=r}};function ei(e,t,n,o){var r,i,a,s;if(t<n){if(s=e.input.slice(t,n),o)for(r=0,i=s.length;r<i;r+=1)9===(a=s.charCodeAt(r))||32<=a&&a<=1114111||en(e,"expected valid JSON character");else q.test(s)&&en(e,"the stream contains non-printable characters");e.result+=s}}function ea(e,t,n,o){var r,i,a,s,l;for("object"==typeof(l=n)&&null!==l||en(e,"cannot merge mappings; the provided source object is unacceptable"),a=0,s=(r=Object.keys(n)).length;a<s;a+=1)i=r[a],D.call(t,i)||(z(t,i,n[i]),o[i]=!0)}function es(e,t,n,o,r,i,a,s,l){var u,c;if(Array.isArray(r))for(u=0,c=(r=Array.prototype.slice.call(r)).length;u<c;u+=1)Array.isArray(r[u])&&en(e,"nested arrays are not supported inside keys"),"object"==typeof r&&"[object Object]"===G(r[u])&&(r[u]="[object Object]");if("object"==typeof r&&"[object Object]"===G(r)&&(r="[object Object]"),r=String(r),null===t&&(t={}),"tag:yaml.org,2002:merge"===o)if(Array.isArray(i))for(u=0,c=i.length;u<c;u+=1)ea(e,t,i[u],n);else ea(e,t,i,n);else!e.json&&!D.call(n,r)&&D.call(t,r)&&(e.line=a||e.line,e.lineStart=s||e.lineStart,e.position=l||e.position,en(e,"duplicated mapping key")),z(t,r,i),delete n[r];return t}function el(e){var t;10===(t=e.input.charCodeAt(e.position))?e.position++:13===t?(e.position++,10===e.input.charCodeAt(e.position)&&e.position++):en(e,"a line break is expected"),e.line+=1,e.lineStart=e.position,e.firstTabInLine=-1}function eu(e,t,n){for(var o=0,r=e.input.charCodeAt(e.position);0!==r;){for(;$(r);)9===r&&-1===e.firstTabInLine&&(e.firstTabInLine=e.position),r=e.input.charCodeAt(++e.position);if(t&&35===r)do r=e.input.charCodeAt(++e.position);while(10!==r&&13!==r&&0!==r)if(B(r))for(el(e),r=e.input.charCodeAt(e.position),o++,e.lineIndent=0;32===r;)e.lineIndent++,r=e.input.charCodeAt(++e.position);else break}return -1!==n&&0!==o&&e.lineIndent<n&&eo(e,"deficient indentation"),o}function ec(e){var t,n=e.position;return!!((45===(t=e.input.charCodeAt(n))||46===t)&&t===e.input.charCodeAt(n+1)&&t===e.input.charCodeAt(n+2)&&(n+=3,0===(t=e.input.charCodeAt(n))||K(t)))||!1}function ep(e,n){1===n?e.result+=" ":n>1&&(e.result+=t("\n",n-1))}function ed(e,t){var n,o,r=e.tag,i=e.anchor,a=[],s=!1;if(-1!==e.firstTabInLine)return!1;for(null!==e.anchor&&(e.anchorMap[e.anchor]=a),o=e.input.charCodeAt(e.position);0!==o&&(-1!==e.firstTabInLine&&(e.position=e.firstTabInLine,en(e,"tab characters must not be used in indentation")),45===o&&K(e.input.charCodeAt(e.position+1)));){if(s=!0,e.position++,eu(e,!0,-1)&&e.lineIndent<=t){a.push(null),o=e.input.charCodeAt(e.position);continue}if(n=e.line,em(e,t,3,!1,!0),a.push(e.result),eu(e,!0,-1),o=e.input.charCodeAt(e.position),(e.line===n||e.lineIndent>t)&&0!==o)en(e,"bad indentation of a sequence entry");else if(e.lineIndent<t)break}return!!s&&(e.tag=r,e.anchor=i,e.kind="sequence",e.result=a,!0)}function em(e,n,o,r,i){var a,s,l,u,c,p,d,m,f,h=1,g=!1,y=!1;if(null!==e.listener&&e.listener("open",e),e.tag=null,e.anchor=null,e.kind=null,e.result=null,a=s=l=4===o||3===o,r&&eu(e,!0,-1)&&(g=!0,e.lineIndent>n?h=1:e.lineIndent===n?h=0:e.lineIndent<n&&(h=-1)),1===h)for(;function(e){var t,n,o,r,i=!1,a=!1;if(33!==(r=e.input.charCodeAt(e.position)))return!1;if(null!==e.tag&&en(e,"duplication of a tag property"),60===(r=e.input.charCodeAt(++e.position))?(i=!0,r=e.input.charCodeAt(++e.position)):33===r?(a=!0,n="!!",r=e.input.charCodeAt(++e.position)):n="!",t=e.position,i){do r=e.input.charCodeAt(++e.position);while(0!==r&&62!==r)e.position<e.length?(o=e.input.slice(t,e.position),r=e.input.charCodeAt(++e.position)):en(e,"unexpected end of the stream within a verbatim tag")}else{for(;0!==r&&!K(r);)33===r&&(a?en(e,"tag suffix cannot contain exclamation marks"):(n=e.input.slice(t-1,e.position+1),W.test(n)||en(e,"named tag handle cannot contain such characters"),a=!0,t=e.position+1)),r=e.input.charCodeAt(++e.position);o=e.input.slice(t,e.position),U.test(o)&&en(e,"tag suffix cannot contain flow indicator characters")}o&&!Y.test(o)&&en(e,"tag name cannot contain such characters: "+o);try{o=decodeURIComponent(o)}catch(t){en(e,"tag name is malformed: "+o)}return i?e.tag=o:D.call(e.tagMap,n)?e.tag=e.tagMap[n]+o:"!"===n?e.tag="!"+o:"!!"===n?e.tag="tag:yaml.org,2002:"+o:en(e,'undeclared tag handle "'+n+'"'),!0}(e)||function(e){var t,n;if(38!==(n=e.input.charCodeAt(e.position)))return!1;for(null!==e.anchor&&en(e,"duplication of an anchor property"),n=e.input.charCodeAt(++e.position),t=e.position;0!==n&&!K(n)&&!J(n);)n=e.input.charCodeAt(++e.position);return e.position===t&&en(e,"name of an anchor node must contain at least one character"),e.anchor=e.input.slice(t,e.position),!0}(e);)eu(e,!0,-1)?(g=!0,l=a,e.lineIndent>n?h=1:e.lineIndent===n?h=0:e.lineIndent<n&&(h=-1)):l=!1;if(l&&(l=g||i),(1===h||4===o)&&(m=1===o||2===o?n:n+1,f=e.position-e.lineStart,1===h?l&&(ed(e,f)||function(e,t,n){var o,r,i,a,s,l,u,c=e.tag,p=e.anchor,d={},m=Object.create(null),f=null,h=null,g=null,y=!1,w=!1;if(-1!==e.firstTabInLine)return!1;for(null!==e.anchor&&(e.anchorMap[e.anchor]=d),u=e.input.charCodeAt(e.position);0!==u;){if(y||-1===e.firstTabInLine||(e.position=e.firstTabInLine,en(e,"tab characters must not be used in indentation")),o=e.input.charCodeAt(e.position+1),i=e.line,(63===u||58===u)&&K(o))63===u?(y&&(es(e,d,m,f,h,null,a,s,l),f=h=g=null),w=!0,y=!0,r=!0):y?(y=!1,r=!0):en(e,"incomplete explicit mapping pair; a key node is missed; or followed by a non-tabulated empty line"),e.position+=1,u=o;else{if(a=e.line,s=e.lineStart,l=e.position,!em(e,n,2,!1,!0))break;if(e.line===i){for(u=e.input.charCodeAt(e.position);$(u);)u=e.input.charCodeAt(++e.position);if(58===u)K(u=e.input.charCodeAt(++e.position))||en(e,"a whitespace character is expected after the key-value separator within a block mapping"),y&&(es(e,d,m,f,h,null,a,s,l),f=h=g=null),w=!0,y=!1,r=!1,f=e.tag,h=e.result;else{if(!w)return e.tag=c,e.anchor=p,!0;en(e,"can not read an implicit mapping pair; a colon is missed")}}else{if(!w)return e.tag=c,e.anchor=p,!0;en(e,"can not read a block mapping entry; a multiline key may not be an implicit key")}}if((e.line===i||e.lineIndent>t)&&(y&&(a=e.line,s=e.lineStart,l=e.position),em(e,t,4,!0,r)&&(y?h=e.result:g=e.result),y||(es(e,d,m,f,h,g,a,s,l),f=h=g=null),eu(e,!0,-1),u=e.input.charCodeAt(e.position)),(e.line===i||e.lineIndent>t)&&0!==u)en(e,"bad indentation of a mapping entry");else if(e.lineIndent<t)break}return y&&es(e,d,m,f,h,null,a,s,l),w&&(e.tag=c,e.anchor=p,e.kind="mapping",e.result=d),w}(e,f,m))||function(e,t){var n,o,r,i,a,s,l,u,c,p,d,m,f=!0,h=e.tag,g=e.anchor,y=Object.create(null);if(91===(m=e.input.charCodeAt(e.position)))a=93,u=!1,i=[];else{if(123!==m)return!1;a=125,u=!0,i={}}for(null!==e.anchor&&(e.anchorMap[e.anchor]=i),m=e.input.charCodeAt(++e.position);0!==m;){if(eu(e,!0,t),(m=e.input.charCodeAt(e.position))===a)return e.position++,e.tag=h,e.anchor=g,e.kind=u?"mapping":"sequence",e.result=i,!0;f?44===m&&en(e,"expected the node content, but found ','"):en(e,"missed comma between flow collection entries"),p=c=d=null,s=l=!1,63===m&&K(e.input.charCodeAt(e.position+1))&&(s=l=!0,e.position++,eu(e,!0,t)),n=e.line,o=e.lineStart,r=e.position,em(e,t,1,!1,!0),p=e.tag,c=e.result,eu(e,!0,t),m=e.input.charCodeAt(e.position),(l||e.line===n)&&58===m&&(s=!0,m=e.input.charCodeAt(++e.position),eu(e,!0,t),em(e,t,1,!1,!0),d=e.result),u?es(e,i,y,p,c,d,n,o,r):s?i.push(es(e,null,y,p,c,d,n,o,r)):i.push(c),eu(e,!0,t),44===(m=e.input.charCodeAt(e.position))?(f=!0,m=e.input.charCodeAt(++e.position)):f=!1}en(e,"unexpected end of the stream within a flow collection")}(e,m)?y=!0:(s&&function(e,n){var o,r,i,a,s,l=1,u=!1,c=!1,p=n,d=0,m=!1;if(124===(s=e.input.charCodeAt(e.position)))i=!1;else{if(62!==s)return!1;i=!0}for(e.kind="scalar",e.result="";0!==s;)if(43===(s=e.input.charCodeAt(++e.position))||45===s)1===l?l=43===s?3:2:en(e,"repeat of a chomping mode identifier");else if((a=48<=(o=s)&&o<=57?o-48:-1)>=0)0===a?en(e,"bad explicit indentation width of a block scalar; it cannot be less than one"):c?en(e,"repeat of an indentation width identifier"):(p=n+a-1,c=!0);else break;if($(s)){do s=e.input.charCodeAt(++e.position);while($(s))if(35===s)do s=e.input.charCodeAt(++e.position);while(!B(s)&&0!==s)}for(;0!==s;){for(el(e),e.lineIndent=0,s=e.input.charCodeAt(e.position);(!c||e.lineIndent<p)&&32===s;)e.lineIndent++,s=e.input.charCodeAt(++e.position);if(!c&&e.lineIndent>p&&(p=e.lineIndent),B(s)){d++;continue}if(e.lineIndent<p){3===l?e.result+=t("\n",u?1+d:d):1===l&&u&&(e.result+="\n");break}for(i?$(s)?(m=!0,e.result+=t("\n",u?1+d:d)):m?(m=!1,e.result+=t("\n",d+1)):0===d?u&&(e.result+=" "):e.result+=t("\n",d):e.result+=t("\n",u?1+d:d),u=!0,c=!0,d=0,r=e.position;!B(s)&&0!==s;)s=e.input.charCodeAt(++e.position);ei(e,r,e.position,!1)}return!0}(e,m)||function(e,t){var n,o,r;if(39!==(n=e.input.charCodeAt(e.position)))return!1;for(e.kind="scalar",e.result="",e.position++,o=r=e.position;0!==(n=e.input.charCodeAt(e.position));)if(39===n){if(ei(e,o,e.position,!0),39!==(n=e.input.charCodeAt(++e.position)))return!0;o=e.position,e.position++,r=e.position}else B(n)?(ei(e,o,r,!0),ep(e,eu(e,!1,t)),o=r=e.position):e.position===e.lineStart&&ec(e)?en(e,"unexpected end of the document within a single quoted scalar"):(e.position++,r=e.position);en(e,"unexpected end of the stream within a single quoted scalar")}(e,m)||function(e,t){var n,o,r,i,a,s,l,u;if(34!==(s=e.input.charCodeAt(e.position)))return!1;for(e.kind="scalar",e.result="",e.position++,n=o=e.position;0!==(s=e.input.charCodeAt(e.position));)if(34===s)return ei(e,n,e.position,!0),e.position++,!0;else if(92===s){if(ei(e,n,e.position,!0),B(s=e.input.charCodeAt(++e.position)))eu(e,!1,t);else if(s<256&&X[s])e.result+=Q[s],e.position++;else if((a=120===(l=s)?2:117===l?4:8*(85===l))>0){for(r=a,i=0;r>0;r--)(a=function(e){var t;return 48<=e&&e<=57?e-48:97<=(t=32|e)&&t<=102?t-97+10:-1}(s=e.input.charCodeAt(++e.position)))>=0?i=(i<<4)+a:en(e,"expected hexadecimal character");e.result+=(u=i)<=65535?String.fromCharCode(u):String.fromCharCode((u-65536>>10)+55296,(u-65536&1023)+56320),e.position++}else en(e,"unknown escape sequence");n=o=e.position}else B(s)?(ei(e,n,o,!0),ep(e,eu(e,!1,t)),n=o=e.position):e.position===e.lineStart&&ec(e)?en(e,"unexpected end of the document within a double quoted scalar"):(e.position++,o=e.position);en(e,"unexpected end of the stream within a double quoted scalar")}(e,m)?y=!0:!function(e){var t,n,o;if(42!==(o=e.input.charCodeAt(e.position)))return!1;for(o=e.input.charCodeAt(++e.position),t=e.position;0!==o&&!K(o)&&!J(o);)o=e.input.charCodeAt(++e.position);return e.position===t&&en(e,"name of an alias node must contain at least one character"),n=e.input.slice(t,e.position),D.call(e.anchorMap,n)||en(e,'unidentified alias "'+n+'"'),e.result=e.anchorMap[n],eu(e,!0,-1),!0}(e)?function(e,t,n){var o,r,i,a,s,l,u,c,p=e.kind,d=e.result;if(K(c=e.input.charCodeAt(e.position))||J(c)||35===c||38===c||42===c||33===c||124===c||62===c||39===c||34===c||37===c||64===c||96===c||(63===c||45===c)&&(K(o=e.input.charCodeAt(e.position+1))||n&&J(o)))return!1;for(e.kind="scalar",e.result="",r=i=e.position,a=!1;0!==c;){if(58===c){if(K(o=e.input.charCodeAt(e.position+1))||n&&J(o))break}else if(35===c){if(K(e.input.charCodeAt(e.position-1)))break}else if(e.position===e.lineStart&&ec(e)||n&&J(c))break;else if(B(c)){if(s=e.line,l=e.lineStart,u=e.lineIndent,eu(e,!1,-1),e.lineIndent>=t){a=!0,c=e.input.charCodeAt(e.position);continue}e.position=i,e.line=s,e.lineStart=l,e.lineIndent=u;break}a&&(ei(e,r,i,!1),ep(e,e.line-s),r=i=e.position,a=!1),$(c)||(i=e.position+1),c=e.input.charCodeAt(++e.position)}return ei(e,r,i,!1),!!e.result||(e.kind=p,e.result=d,!1)}(e,m,1===o)&&(y=!0,null===e.tag&&(e.tag="?")):(y=!0,(null!==e.tag||null!==e.anchor)&&en(e,"alias node should not have any properties")),null!==e.anchor&&(e.anchorMap[e.anchor]=e.result)):0===h&&(y=l&&ed(e,f))),null===e.tag)null!==e.anchor&&(e.anchorMap[e.anchor]=e.result);else if("?"===e.tag){for(null!==e.result&&"scalar"!==e.kind&&en(e,'unacceptable node kind for !<?> tag; it should be "scalar", not "'+e.kind+'"'),u=0,c=e.implicitTypes.length;u<c;u+=1)if((d=e.implicitTypes[u]).resolve(e.result)){e.result=d.construct(e.result),e.tag=d.tag,null!==e.anchor&&(e.anchorMap[e.anchor]=e.result);break}}else if("!"!==e.tag){if(D.call(e.typeMap[e.kind||"fallback"],e.tag))d=e.typeMap[e.kind||"fallback"][e.tag];else for(u=0,d=null,c=(p=e.typeMap.multi[e.kind||"fallback"]).length;u<c;u+=1)if(e.tag.slice(0,p[u].tag.length)===p[u].tag){d=p[u];break}d||en(e,"unknown tag !<"+e.tag+">"),null!==e.result&&d.kind!==e.kind&&en(e,"unacceptable node kind for !<"+e.tag+'> tag; it should be "'+d.kind+'", not "'+e.kind+'"'),d.resolve(e.result,e.tag)?(e.result=d.construct(e.result,e.tag),null!==e.anchor&&(e.anchorMap[e.anchor]=e.result)):en(e,"cannot resolve a node with !<"+e.tag+"> explicit tag")}return null!==e.listener&&e.listener("close",e),null!==e.tag||null!==e.anchor||y}function ef(e,t){e=String(e),t=t||{},0!==e.length&&(10!==e.charCodeAt(e.length-1)&&13!==e.charCodeAt(e.length-1)&&(e+="\n"),65279===e.charCodeAt(0)&&(e=e.slice(1)));var n=new ee(e,t),o=e.indexOf("\0");for(-1!==o&&(n.position=o,en(n,"null byte is not allowed in input")),n.input+="\0";32===n.input.charCodeAt(n.position);)n.lineIndent+=1,n.position+=1;for(;n.position<n.length-1;)!function(e){var t,n,o,r,i=e.position,a=!1;for(e.version=null,e.checkLineBreaks=e.legacy,e.tagMap=Object.create(null),e.anchorMap=Object.create(null);0!==(r=e.input.charCodeAt(e.position))&&(eu(e,!0,-1),r=e.input.charCodeAt(e.position),!(e.lineIndent>0)&&37===r);){for(a=!0,r=e.input.charCodeAt(++e.position),t=e.position;0!==r&&!K(r);)r=e.input.charCodeAt(++e.position);for(n=e.input.slice(t,e.position),o=[],n.length<1&&en(e,"directive name must not be less than one character in length");0!==r;){for(;$(r);)r=e.input.charCodeAt(++e.position);if(35===r){do r=e.input.charCodeAt(++e.position);while(0!==r&&!B(r))break}if(B(r))break;for(t=e.position;0!==r&&!K(r);)r=e.input.charCodeAt(++e.position);o.push(e.input.slice(t,e.position))}0!==r&&el(e),D.call(er,n)?er[n](e,n,o):eo(e,'unknown document directive "'+n+'"')}if(eu(e,!0,-1),0===e.lineIndent&&45===e.input.charCodeAt(e.position)&&45===e.input.charCodeAt(e.position+1)&&45===e.input.charCodeAt(e.position+2)?(e.position+=3,eu(e,!0,-1)):a&&en(e,"directives end mark is expected"),em(e,e.lineIndent-1,4,!1,!0),eu(e,!0,-1),e.checkLineBreaks&&H.test(e.input.slice(i,e.position))&&eo(e,"non-ASCII line breaks are interpreted as content"),e.documents.push(e.result),e.position===e.lineStart&&ec(e)){46===e.input.charCodeAt(e.position)&&(e.position+=3,eu(e,!0,-1));return}e.position<e.length-1&&en(e,"end of the stream or a document separator is expected")}(n);return n.documents}var eh=Object.prototype.toString,eg=Object.prototype.hasOwnProperty,ey={};ey[0]="\\0",ey[7]="\\a",ey[8]="\\b",ey[9]="\\t",ey[10]="\\n",ey[11]="\\v",ey[12]="\\f",ey[13]="\\r",ey[27]="\\e",ey[34]='\\"',ey[92]="\\\\",ey[133]="\\N",ey[160]="\\_",ey[8232]="\\L",ey[8233]="\\P";var ew=["y","Y","yes","Yes","YES","on","On","ON","n","N","no","No","NO","off","Off","OFF"],eb=/^[-+]?[0-9_]+(?::[0-9_]+)+(?:\.[0-9_]*)?$/;function e_(e){this.schema=e.schema||N,this.indent=Math.max(1,e.indent||2),this.noArrayIndent=e.noArrayIndent||!1,this.skipInvalid=e.skipInvalid||!1,this.flowLevel=null==e.flowLevel?-1:e.flowLevel,this.styleMap=function(e,t){var n,o,r,i,a,s,l;if(null===t)return{};for(r=0,n={},i=(o=Object.keys(t)).length;r<i;r+=1)s=String(t[a=o[r]]),"!!"===a.slice(0,2)&&(a="tag:yaml.org,2002:"+a.slice(2)),(l=e.compiledTypeMap.fallback[a])&&eg.call(l.styleAliases,s)&&(s=l.styleAliases[s]),n[a]=s;return n}(this.schema,e.styles||null),this.sortKeys=e.sortKeys||!1,this.lineWidth=e.lineWidth||80,this.noRefs=e.noRefs||!1,this.noCompatMode=e.noCompatMode||!1,this.condenseFlow=e.condenseFlow||!1,this.quotingType='"'===e.quotingType?2:1,this.forceQuotes=e.forceQuotes||!1,this.replacer="function"==typeof e.replacer?e.replacer:null,this.implicitTypes=this.schema.compiledImplicit,this.explicitTypes=this.schema.compiledExplicit,this.tag=null,this.result="",this.duplicates=[],this.usedDuplicates=null}function ev(e,n){for(var o,r=t(" ",n),i=0,a=-1,s="",l=e.length;i<l;)-1===(a=e.indexOf("\n",i))?(o=e.slice(i),i=l):(o=e.slice(i,a+1),i=a+1),o.length&&"\n"!==o&&(s+=r),s+=o;return s}function ek(e,n){return"\n"+t(" ",e.indent*n)}function ex(e){return 32===e||9===e}function eS(e){return 32<=e&&e<=126||161<=e&&e<=55295&&8232!==e&&8233!==e||57344<=e&&e<=65533&&65279!==e||65536<=e&&e<=1114111}function eA(e){return eS(e)&&65279!==e&&13!==e&&10!==e}function eO(e,t,n){var o=eA(e),r=o&&!ex(e);return(n?o:o&&44!==e&&91!==e&&93!==e&&123!==e&&125!==e)&&35!==e&&!(58===t&&!r)||eA(t)&&!ex(t)&&35===e||58===t&&r}function eC(e,t){var n,o=e.charCodeAt(t);return o>=55296&&o<=56319&&t+1<e.length&&(n=e.charCodeAt(t+1))>=56320&&n<=57343?(o-55296)*1024+n-56320+65536:o}function eT(e){return/^\n* /.test(e)}function eE(e,t){var n=eT(e)?String(t):"",o="\n"===e[e.length-1];return n+(o&&("\n"===e[e.length-2]||"\n"===e)?"+":o?"":"-")+"\n"}function eI(e){return"\n"===e[e.length-1]?e.slice(0,-1):e}function eF(e,t){if(""===e||" "===e[0])return e;for(var n,o,r=/ [^ ]/g,i=0,a=0,s=0,l="";n=r.exec(e);)(s=n.index)-i>t&&(o=a>i?a:s,l+="\n"+e.slice(i,o),i=o+1),a=s;return l+="\n",e.length-i>t&&a>i?l+=e.slice(i,a)+"\n"+e.slice(a+1):l+=e.slice(i),l.slice(1)}function eM(e,t,n,o){var r,i,a,s="",l=e.tag;for(r=0,i=n.length;r<i;r+=1)a=n[r],e.replacer&&(a=e.replacer.call(n,String(r),a)),(eR(e,t+1,a,!0,!0,!1,!0)||void 0===a&&eR(e,t+1,null,!0,!0,!1,!0))&&(o&&""===s||(s+=ek(e,t)),e.dump&&10===e.dump.charCodeAt(0)?s+="-":s+="- ",s+=e.dump);e.tag=l,e.dump=s||"[]"}function eL(e,t,n){var o,i,a,s,l,u;for(a=0,s=(i=n?e.explicitTypes:e.implicitTypes).length;a<s;a+=1)if(((l=i[a]).instanceOf||l.predicate)&&(!l.instanceOf||"object"==typeof t&&t instanceof l.instanceOf)&&(!l.predicate||l.predicate(t))){if(n?l.multi&&l.representName?e.tag=l.representName(t):e.tag=l.tag:e.tag="?",l.represent){if(u=e.styleMap[l.tag]||l.defaultStyle,"[object Function]"===eh.call(l.represent))o=l.represent(t,u);else if(eg.call(l.represent,u))o=l.represent[u](t,u);else throw new r("!<"+l.tag+'> tag resolver accepts not "'+u+'" style');e.dump=o}return!0}return!1}function eR(e,n,o,i,a,s,l){e.tag=null,e.dump=o,eL(e,o,!1)||eL(e,o,!0);var u,c=eh.call(e.dump),p=i;i&&(i=e.flowLevel<0||e.flowLevel>n);var d,m,f,h="[object Object]"===c||"[object Array]"===c;if(h&&(f=-1!==(m=e.duplicates.indexOf(o))),(null!==e.tag&&"?"!==e.tag||f||2!==e.indent&&n>0)&&(a=!1),f&&e.usedDuplicates[m])e.dump="*ref_"+m;else{if(h&&f&&!e.usedDuplicates[m]&&(e.usedDuplicates[m]=!0),"[object Object]"===c)i&&0!==Object.keys(e.dump).length?(!function(e,t,n,o){var i,a,s,l,u,c,p="",d=e.tag,m=Object.keys(n);if(!0===e.sortKeys)m.sort();else if("function"==typeof e.sortKeys)m.sort(e.sortKeys);else if(e.sortKeys)throw new r("sortKeys must be a boolean or a function");for(i=0,a=m.length;i<a;i+=1)c="",o&&""===p||(c+=ek(e,t)),l=n[s=m[i]],e.replacer&&(l=e.replacer.call(n,s,l)),eR(e,t+1,s,!0,!0,!0)&&((u=null!==e.tag&&"?"!==e.tag||e.dump&&e.dump.length>1024)&&(e.dump&&10===e.dump.charCodeAt(0)?c+="?":c+="? "),c+=e.dump,u&&(c+=ek(e,t)),eR(e,t+1,l,!0,u)&&(e.dump&&10===e.dump.charCodeAt(0)?c+=":":c+=": ",c+=e.dump,p+=c));e.tag=d,e.dump=p||"{}"}(e,n,e.dump,a),f&&(e.dump="&ref_"+m+e.dump)):(!function(e,t,n){var o,r,i,a,s,l="",u=e.tag,c=Object.keys(n);for(o=0,r=c.length;o<r;o+=1)s="",""!==l&&(s+=", "),e.condenseFlow&&(s+='"'),a=n[i=c[o]],e.replacer&&(a=e.replacer.call(n,i,a)),eR(e,t,i,!1,!1)&&(e.dump.length>1024&&(s+="? "),s+=e.dump+(e.condenseFlow?'"':"")+":"+(e.condenseFlow?"":" "),eR(e,t,a,!1,!1)&&(s+=e.dump,l+=s));e.tag=u,e.dump="{"+l+"}"}(e,n,e.dump),f&&(e.dump="&ref_"+m+" "+e.dump));else if("[object Array]"===c)i&&0!==e.dump.length?(e.noArrayIndent&&!l&&n>0?eM(e,n-1,e.dump,a):eM(e,n,e.dump,a),f&&(e.dump="&ref_"+m+e.dump)):(!function(e,t,n){var o,r,i,a="",s=e.tag;for(o=0,r=n.length;o<r;o+=1)i=n[o],e.replacer&&(i=e.replacer.call(n,String(o),i)),(eR(e,t,i,!1,!1)||void 0===i&&eR(e,t,null,!1,!1))&&(""!==a&&(a+=","+(e.condenseFlow?"":" ")),a+=e.dump);e.tag=s,e.dump="["+a+"]"}(e,n,e.dump),f&&(e.dump="&ref_"+m+" "+e.dump));else if("[object String]"===c)"?"!==e.tag&&(u=e.dump,e.dump=function(){if(0===u.length)return 2===e.quotingType?'""':"''";if(!e.noCompatMode&&(-1!==ew.indexOf(u)||eb.test(u)))return 2===e.quotingType?'"'+u+'"':"'"+u+"'";var o=e.indent*Math.max(1,n),i=-1===e.lineWidth?-1:Math.max(Math.min(e.lineWidth,40),e.lineWidth-o);switch(function(e,t,n,o,r,i,a,s){var l,u,c,p=0,d=null,m=!1,f=!1,h=-1!==o,g=-1,y=eS(l=eC(e,0))&&65279!==l&&!ex(l)&&45!==l&&63!==l&&58!==l&&44!==l&&91!==l&&93!==l&&123!==l&&125!==l&&35!==l&&38!==l&&42!==l&&33!==l&&124!==l&&61!==l&&62!==l&&39!==l&&34!==l&&37!==l&&64!==l&&96!==l&&!ex(u=eC(e,e.length-1))&&58!==u;if(t||a)for(c=0;c<e.length;p>=65536?c+=2:c++){if(!eS(p=eC(e,c)))return 5;y=y&&eO(p,d,s),d=p}else{for(c=0;c<e.length;p>=65536?c+=2:c++){if(10===(p=eC(e,c)))m=!0,h&&(f=f||c-g-1>o&&" "!==e[g+1],g=c);else if(!eS(p))return 5;y=y&&eO(p,d,s),d=p}f=f||h&&c-g-1>o&&" "!==e[g+1]}return m||f?n>9&&eT(e)?5:a?2===i?5:2:f?4:3:!y||a||r(e)?2===i?5:2:1}(u,s||e.flowLevel>-1&&n>=e.flowLevel,e.indent,i,function(t){var n,o;for(n=0,o=e.implicitTypes.length;n<o;n+=1)if(e.implicitTypes[n].resolve(t))return!0;return!1},e.quotingType,e.forceQuotes&&!s,p)){case 1:return u;case 2:return"'"+u.replace(/'/g,"''")+"'";case 3:return"|"+eE(u,e.indent)+eI(ev(u,o));case 4:return">"+eE(u,e.indent)+eI(ev(function(e,t){for(var n,o,r,i=/(\n+)([^\n]*)/g,a=(i.lastIndex=n=-1!==(n=e.indexOf("\n"))?n:e.length,eF(e.slice(0,n),t)),s="\n"===e[0]||" "===e[0];r=i.exec(e);){var l=r[1],u=r[2];o=" "===u[0],a+=l+(s||o||""===u?"":"\n")+eF(u,t),s=o}return a}(u,i),o));case 5:return'"'+function(e){for(var n,o="",i=0,a=0;a<e.length;i>=65536?a+=2:a++)!(n=ey[i=eC(e,a)])&&eS(i)?(o+=e[a],i>=65536&&(o+=e[a+1])):o+=n||function(e){var n,o,i;if(n=e.toString(16).toUpperCase(),e<=255)o="x",i=2;else if(e<=65535)o="u",i=4;else if(e<=0xffffffff)o="U",i=8;else throw new r("code point within a string may not be greater than 0xFFFFFFFF");return"\\"+o+t("0",i-n.length)+n}(i);return o}(u)+'"';default:throw new r("impossible error: invalid scalar style")}}());else{if("[object Undefined]"===c||e.skipInvalid)return!1;throw new r("unacceptable kind of an object to dump "+c)}null!==e.tag&&"?"!==e.tag&&(d=encodeURI("!"===e.tag[0]?e.tag.slice(1):e.tag).replace(/!/g,"%21"),d="!"===e.tag[0]?"!"+d:"tag:yaml.org,2002:"===d.slice(0,18)?"!!"+d.slice(18):"!<"+d+">",e.dump=d+" "+e.dump)}return!0}function eP(e,t){return function(){throw Error("Function yaml."+e+" is removed in js-yaml 4. Use yaml."+t+" instead, which is now safe by default.")}}var ej={Type:c,Schema:d,FAILSAFE_SCHEMA:g,JSON_SCHEMA:x,CORE_SCHEMA:x,DEFAULT_SCHEMA:N,load:function(e,t){var n=ef(e,t);if(0!==n.length){if(1===n.length)return n[0];throw new r("expected a single document in the stream, but found more")}},loadAll:function(e,t,n){null!==t&&"object"==typeof t&&void 0===n&&(n=t,t=null);var o=ef(e,n);if("function"!=typeof t)return o;for(var r=0,i=o.length;r<i;r+=1)t(o[r])},dump:function(e,t){var n=new e_(t=t||{});n.noRefs||function(e,t){var n,o,r=[],i=[];for(function e(t,n,o){var r,i,a;if(null!==t&&"object"==typeof t)if(-1!==(i=n.indexOf(t)))-1===o.indexOf(i)&&o.push(i);else if(n.push(t),Array.isArray(t))for(i=0,a=t.length;i<a;i+=1)e(t[i],n,o);else for(i=0,a=(r=Object.keys(t)).length;i<a;i+=1)e(t[r[i]],n,o)}(e,r,i),n=0,o=i.length;n<o;n+=1)t.duplicates.push(r[i[n]]);t.usedDuplicates=Array(o)}(e,n);var o=e;return(n.replacer&&(o=n.replacer.call({"":o},"",o)),eR(n,0,o,!0,!0))?n.dump+"\n":""},YAMLException:r,types:{binary:E,float:k,map:h,null:y,pairs:R,set:j,timestamp:O,bool:w,int:b,merge:C,omap:M,seq:f,str:m},safeLoad:eP("safeLoad","load"),safeLoadAll:eP("safeLoadAll","loadAll"),safeDump:eP("safeDump","dump")};e.s(["default",()=>ej])},37364,e=>{"use strict";let t={"test-complex-docker-workflow":`name: test-complex-docker-workflow
kind: module
description: Complex workflow demonstrating bash, function steps with docker step_runner

params:
  - name: target
    required: true
  - name: output_dir
    default: /tmp/osm-complex-test
  - name: threads
    default: "5"

steps:
  # Step 1: Setup - Create directories using function
  - name: setup-workspace
    type: function
    log: "Setting up workspace for {{target}}"
    function: createDir("{{output_dir}}")
    exports:
      workspace_created: "output"

  # Step 2: Create input file with bash
  - name: create-target-list
    type: bash
    log: "Creating target list for {{target}}"
    commands:
      - mkdir -p {{output_dir}}/targets
      - |
        cat > {{output_dir}}/targets/hosts.txt << 'EOF'
        sub1.{{target}}
        sub2.{{target}}
        api.{{target}}
        www.{{target}}
        admin.{{target}}
        EOF
    exports:
      target_file: "{{output_dir}}/targets/hosts.txt"

  # Step 3: Docker-based DNS resolution simulation
  - name: dns-resolve
    type: remote-bash
    log: "Resolving DNS for targets in Docker"
    timeout: 60
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      env:
        TARGET_DOMAIN: "{{target}}"
      volumes:
        - "{{output_dir}}:/workspace"
      workdir: /workspace
    command: |
      echo "Resolving DNS for $TARGET_DOMAIN"
      cat /workspace/targets/hosts.txt | while read host; do
        echo "$host -> 127.0.0.1" >> /workspace/dns-resolved.txt
      done
      echo "DNS resolution complete"
    exports:
      dns_output: "{{output_dir}}/dns-resolved.txt"

  # Step 4: Parallel docker commands - simulating port scanning
  - name: parallel-port-scan
    type: remote-bash
    log: "Running parallel port scans in Docker"
    timeout: 120
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    parallel_commands:
      - 'echo "Scanning ports 1-1000 on {{target}}" && sleep 1 && echo "Port 80 open" > /workspace/ports-1.txt'
      - 'echo "Scanning ports 1001-2000 on {{target}}" && sleep 1 && echo "Port 443 open" > /workspace/ports-2.txt'
      - 'echo "Scanning ports 2001-3000 on {{target}}" && sleep 1 && echo "Port 8080 open" > /workspace/ports-3.txt'
      - 'echo "Scanning ports 3001-4000 on {{target}}" && sleep 1 && echo "Port 3306 open" > /workspace/ports-4.txt'

  # Step 5: Merge port scan results
  - name: merge-port-results
    type: bash
    log: "Merging port scan results"
    command: cat {{output_dir}}/ports-*.txt > {{output_dir}}/all-ports.txt
    exports:
      ports_file: "{{output_dir}}/all-ports.txt"

  # Step 6: Function to check file existence
  - name: verify-ports-file
    type: function
    log: "Verifying ports file exists"
    function: fileExists("{{ports_file}}")
    exports:
      ports_verified: "output"

  # Step 7: Docker-based HTTP probing with parallel steps
  - name: http-probe-parallel
    type: parallel-steps
    log: "Running parallel HTTP probes"
    parallel_steps:
      - name: probe-http
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTP on port 80"
          echo "http://{{target}}:80 [200]" > /workspace/http-80.txt
      - name: probe-https
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTPS on port 443"
          echo "https://{{target}}:443 [200]" > /workspace/https-443.txt
      - name: probe-alt
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing alternate port 8080"
          echo "http://{{target}}:8080 [404]" > /workspace/http-8080.txt

  # Step 8: Foreach loop with docker - process each subdomain
  - name: process-subdomains
    type: foreach
    log: "Processing each subdomain"
    input: "{{output_dir}}/targets/hosts.txt"
    variable: subdomain
    threads: 3
    step:
      name: scan-subdomain
      type: remote-bash
      step_runner: docker
      step_runner_config:
        image: alpine:latest
        volumes:
          - "{{output_dir}}:/workspace"
      command: |
        echo "Scanning [[subdomain]]..."
        echo "[[subdomain]]: status=200, title=Example" >> /workspace/subdomain-results.txt

  # Step 9: Read results with function
  - name: read-subdomain-results
    type: function
    log: "Reading subdomain scan results"
    function: readFile("{{output_dir}}/subdomain-results.txt")
    exports:
      scan_results: "output"

  # Step 10: Decision based routing
  - name: check-results
    type: bash
    log: "Checking scan results"
    command: wc -l < {{output_dir}}/subdomain-results.txt
    exports:
      result_count: "output"
    decision:
      - condition: result_count == "0"
        next: "_end"
      - condition: result_count != "0"
        next: "generate-report"

  # Step 11: Generate final report in docker
  - name: generate-report
    type: remote-bash
    log: "Generating final report"
    timeout: 30
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    commands:
      - echo "=== Scan Report for {{target}} ===" > /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- DNS Results ---" >> /workspace/report.txt
      - cat /workspace/dns-resolved.txt >> /workspace/report.txt 2>/dev/null || echo "No DNS results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Open Ports ---" >> /workspace/report.txt
      - cat /workspace/all-ports.txt >> /workspace/report.txt 2>/dev/null || echo "No ports found" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Subdomain Results ---" >> /workspace/report.txt
      - cat /workspace/subdomain-results.txt >> /workspace/report.txt 2>/dev/null || echo "No subdomain results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "Report generated at $(date)" >> /workspace/report.txt
    exports:
      report_file: "{{output_dir}}/report.txt"

  # Step 12: Parallel functions to get file stats
  - name: get-file-stats
    type: function
    log: "Getting file statistics"
    parallel_functions:
      - fileLength("{{output_dir}}/report.txt")
      - fileExists("{{output_dir}}/all-ports.txt")
      - trim("  {{target}}  ")
    exports:
      file_stats: "output"

  # Step 13: Cleanup (optional - controlled by pre_condition)
  - name: cleanup-temp-files
    type: bash
    log: "Cleaning up temporary files"
    pre_condition: "false"
    command: rm -rf {{output_dir}}/ports-*.txt
    on_error:
      - action: log
        message: "Cleanup failed but continuing"
      - action: continue
`,"test-decision":`name: test-decision
kind: module
description: Test conditional step routing with decision

params:
  - name: target
    required: true

steps:
  - name: check-condition
    type: bash
    command: echo "{{target}}"
    exports:
      target_value: "output"
    decision:
      - condition: target_value == "skip"
        next: "_end"
      - condition: target_value == "jump"
        next: "final-step"

  - name: middle-step
    type: bash
    command: echo "middle executed"
    exports:
      middle_output: "output"

  - name: final-step
    type: bash
    command: echo "final executed"
    exports:
      final_output: "output"
`,"test-docker-flow":`name: test-docker-flow
kind: flow
description: Flow orchestrating multiple Docker-based security scanning modules

params:
  - name: target
    required: true
  - name: Output
    default: /tmp/osm-docker-flow
  - name: mode
    default: "full"
  - name: threads
    default: "10"
  - name: skip_vuln_scan
    default: "false"

modules:
  # Module 1: Initial reconnaissance
  - name: recon-module
    path: modules/test-docker-recon
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/recon"
      threads: "{{threads}}"
    on_success:
      - action: log
        message: "Reconnaissance completed for {{target}}"
      - action: export
        key: recon_complete
        value: "true"
    on_error:
      - action: log
        message: "Reconnaissance failed for {{target}}"
      - action: abort

  # Module 2: Subdomain enumeration (depends on recon)
  - name: subdomain-module
    path: modules/test-docker-subdomain
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/subdomains"
      wordlist: "/usr/share/wordlists/subdomains.txt"
    condition: "mode == 'full' || mode == 'subdomain'"
    on_success:
      - action: export
        key: subdomains_file
        value: "{{Output}}/subdomains/all.txt"

  # Module 3: Port scanning (parallel with subdomain)
  - name: portscan-module
    path: modules/test-docker-portscan
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/ports"
      port_range: "1-10000"
      rate: "1000"
    condition: "mode == 'full' || mode == 'portscan'"

  # Module 4: HTTP probing (depends on subdomain results)
  - name: httpx-module
    path: modules/test-docker-httpx
    depends_on:
      - subdomain-module
    params:
      input: "{{subdomains_file}}"
      output_dir: "{{Output}}/http"
      threads: "{{threads}}"
    on_success:
      - action: export
        key: alive_hosts
        value: "{{Output}}/http/alive.txt"
      - action: export
        key: httpx_json
        value: "{{Output}}/http/httpx.json"
    decision:
      - condition: "fileLength('{{Output}}/http/alive.txt') == 0"
        next: "report-module"

  # Module 5: Technology detection (depends on HTTP probe)
  - name: tech-detect-module
    path: modules/test-docker-techdetect
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/tech"

  # Module 6: Screenshot capture (parallel with tech detection)
  - name: screenshot-module
    path: modules/test-docker-screenshot
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/screenshots"
      threads: "5"

  # Module 7: Vulnerability scanning (conditional)
  - name: vulnscan-module
    path: modules/test-docker-scanning
    depends_on:
      - httpx-module
      - tech-detect-module
    params:
      target: "{{target}}"
      Output: "{{Output}}/vulns"
      severity: "critical,high,medium"
      threads: "{{threads}}"
    condition: "skip_vuln_scan != 'true'"
    on_error:
      - action: log
        message: "Vulnerability scan encountered errors but continuing"
      - action: continue

  # Module 8: Directory bruteforcing (optional - depends on mode)
  - name: dirbrute-module
    path: modules/test-docker-dirbrute
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/dirs"
      wordlist: "/usr/share/wordlists/common.txt"
      threads: "20"
    condition: "mode == 'full'"

  # Module 9: JavaScript analysis (depends on dir results)
  - name: js-analysis-module
    path: modules/test-docker-jsanalysis
    depends_on:
      - dirbrute-module
    params:
      input: "{{Output}}/dirs/js-files.txt"
      output_dir: "{{Output}}/js"
    condition: "mode == 'full'"

  # Module 10: Final report generation
  - name: report-module
    path: modules/test-docker-report
    depends_on:
      - screenshot-module
      - vulnscan-module
      - tech-detect-module
    params:
      target: "{{target}}"
      input_dir: "{{Output}}"
      output_dir: "{{Output}}/reports"
      format: "html,json,markdown"
    on_success:
      - action: log
        message: "Flow completed successfully for {{target}}"
      - action: notify
        message: "Security assessment complete: {{target}}"
`,"test-loop":`name: test-loop
kind: module
description: Test foreach loop with threading

params:
  - name: target
    required: true

steps:
  - name: create-input
    type: bash
    commands:
      - mkdir -p {{Output}}
      - printf 'one\\ntwo\\nthree\\nfour\\nfive\\n' > {{Output}}/items.txt

  - name: process-items
    type: foreach
    input: "{{Output}}/items.txt"
    variable: item
    threads: 2
    step:
      name: process-item
      type: bash
      command: echo "Processing [[item]] for {{target}}"
`,"comprehensive-flow-example":`# =============================================================================
# Flow Workflow: Comprehensive Example
# =============================================================================
# This file demonstrates ALL fields available in a flow-kind workflow.
# Flows orchestrate multiple modules with dependencies, conditions, and routing.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# Same as module workflows (kind, name, description, tags, params, etc.)
# -----------------------------------------------------------------------------

# kind: Workflow type - "flow" orchestrates multiple modules
kind: flow

# name: Unique identifier for this workflow (required)
name: comprehensive-flow-example

# description: Human-readable description
description: Demonstrates all flow-specific fields including modules, dependencies, conditions, and decisions

# tags: Comma-separated tags for filtering
tags: flow, comprehensive, example

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Parameters available to all modules in this flow
# -----------------------------------------------------------------------------
params:
  - name: threads
    default: "10"

  - name: timeout
    default: "3600"

  - name: scan_depth
    default: "normal"

  - name: output_format
    default: "json"

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Flow-level dependencies checked before any module executes
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - nmap
    - nuclei
    - httpx

  files:
    - /tmp

  variables:
    - name: Target
      type: domain
      required: true

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Reports aggregated from all modules in this flow
# -----------------------------------------------------------------------------
reports:
  - name: flow-summary
    path: "{{Output}}/flow-summary.json"
    type: json
    description: Aggregated results from all modules

  - name: vulnerabilities
    path: "{{Output}}/vulnerabilities.txt"
    type: text
    description: All discovered vulnerabilities

# -----------------------------------------------------------------------------
# PREFERENCES SECTION
# Flow-level preferences apply to all module executions
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: false
  heuristics_check: 'basic'

# -----------------------------------------------------------------------------
# MODULES SECTION (Flow-specific)
# Ordered list of module references to execute
# =============================================================================
modules:
  # ===========================================================================
  # Module Reference: Basic Configuration
  # ===========================================================================
  - # name: Display name for this module execution (required)
    name: reconnaissance

    # path: Path to the module YAML file (required)
    # Can be relative to workflows directory or absolute
    path: modules/recon.yaml

    # params: Parameters to pass to this module
    # Overrides module defaults and flow-level params
    params:
      threads: "20"  # Override flow-level threads
      output_dir: "{{Output}}/recon"

  # ===========================================================================
  # Module Reference: With Dependencies (depends_on)
  # ===========================================================================
  - name: port-scanning
    path: modules/portscan.yaml

    # depends_on: List of module names that must complete before this module runs
    # Creates a DAG (Directed Acyclic Graph) for execution order
    depends_on:
      - reconnaissance

    params:
      target_list: "{{Output}}/recon/subdomains.txt"
      threads: "{{threads}}"

  # ===========================================================================
  # Module Reference: With Condition
  # ===========================================================================
  - name: web-scanning
    path: modules/webscan.yaml

    depends_on:
      - port-scanning

    # condition: JavaScript expression - module only runs if evaluates to true
    # Can reference exported variables from previous modules
    condition: 'fileLength("{{Output}}/portscan/http-services.txt") > 0'

    params:
      input: "{{Output}}/portscan/http-services.txt"

  # ===========================================================================
  # Module Reference: With on_success Handler
  # ===========================================================================
  - name: vulnerability-scanning
    path: modules/vuln-scan.yaml

    depends_on:
      - web-scanning

    condition: 'fileExists("{{Output}}/webscan/endpoints.txt")'

    params:
      endpoints: "{{Output}}/webscan/endpoints.txt"
      timeout: "{{timeout}}"

    # on_success: Actions to execute when this module completes successfully
    on_success:
      # action: log - Log a message
      - action: log
        message: "Vulnerability scanning completed for {{Target}}"

      # action: export - Export a variable for subsequent modules
      - action: export
        name: vuln_scan_complete
        value: "true"

      # action: notify - Send a notification
      - action: notify
        notify: "Vulnerability scan finished for {{Target}}"

      # action: run - Execute a follow-up step
      - action: run
        type: bash
        command: 'echo "Vuln scan done" >> {{Output}}/flow-log.txt'

      # action: run with functions
      - action: run
        type: function
        functions:
          - 'log_info("Module completed successfully")'

  # ===========================================================================
  # Module Reference: With on_error Handler
  # ===========================================================================
  - name: exploit-verification
    path: modules/exploit-verify.yaml

    depends_on:
      - vulnerability-scanning

    condition: '{{vuln_scan_complete}} == "true"'

    params:
      vulns_file: "{{Output}}/vuln-scan/vulnerabilities.json"

    # on_error: Actions to execute when this module fails
    on_error:
      # action: log - Log error message
      - action: log
        message: "Exploit verification failed for {{Target}}"
        # condition: Only execute if this condition is true
        condition: 'true'

      # action: continue - Allow flow to continue despite error
      - action: continue
        message: "Continuing flow despite exploit verification failure"

      # action: abort - Stop the entire flow
      # (Usually with a condition so it doesn't always abort)
      - action: abort
        message: "Critical failure - aborting flow"
        condition: 'false'  # Only abort under specific conditions

      # action: notify - Alert on failure
      - action: notify
        notify: "Module failed: exploit-verification for {{Target}}"

      # action: export - Export error state
      - action: export
        name: exploit_verify_failed
        value: "true"

  # ===========================================================================
  # Module Reference: With Decision Routing
  # ===========================================================================
  - name: deep-scan
    path: modules/deep-scan.yaml

    depends_on:
      - vulnerability-scanning

    # decision: Conditional routing based on results
    # Determines which module to execute next based on conditions
    decision:
      # condition: JavaScript expression to evaluate
      # next: Module name to jump to, or "_end" to finish flow
      - condition: 'fileLength("{{Output}}/vuln-scan/critical.txt") > 0'
        next: notification-critical

      - condition: 'fileLength("{{Output}}/vuln-scan/high.txt") > 0'
        next: notification-high

      # Default case - continue to next module in list
      - condition: 'true'
        next: cleanup

    params:
      scan_depth: "{{scan_depth}}"

  # ===========================================================================
  # Module Reference: Notification branches (targets of decision routing)
  # ===========================================================================
  - name: notification-critical
    path: modules/notify.yaml

    # Note: This module can be jumped to via decision routing
    # It won't run in normal sequential flow unless explicitly in depends_on

    params:
      severity: critical
      message: "Critical vulnerabilities found for {{Target}}"
      channel: security-alerts

    on_success:
      - action: export
        name: notification_sent
        value: "critical"

  - name: notification-high
    path: modules/notify.yaml

    params:
      severity: high
      message: "High severity vulnerabilities found for {{Target}}"
      channel: security-team

    on_success:
      - action: export
        name: notification_sent
        value: "high"

  # ===========================================================================
  # Module Reference: Parallel Module Execution
  # Modules with same depends_on and no inter-dependencies run in parallel
  # ===========================================================================
  - name: ssl-analysis
    path: modules/ssl-check.yaml

    depends_on:
      - port-scanning  # Same dependency as web-scanning

    params:
      input: "{{Output}}/portscan/ssl-services.txt"

  - name: dns-analysis
    path: modules/dns-check.yaml

    depends_on:
      - reconnaissance  # Can run in parallel with port-scanning

    params:
      domains: "{{Output}}/recon/subdomains.txt"

  # ===========================================================================
  # Module Reference: Cleanup/Final Module
  # ===========================================================================
  - name: cleanup
    path: modules/cleanup.yaml

    # depends_on multiple modules - waits for all to complete
    depends_on:
      - vulnerability-scanning
      - exploit-verification
      - ssl-analysis
      - dns-analysis

    # condition with multiple checks
    condition: 'true'  # Always run cleanup

    params:
      output_dir: "{{Output}}"
      format: "{{output_format}}"

    on_success:
      - action: log
        message: "Flow completed successfully for {{Target}}"

      - action: notify
        notify: "Security scan flow completed for {{Target}}"

      - action: export
        name: flow_status
        value: "completed"

    on_error:
      - action: log
        message: "Cleanup failed but flow results are preserved"

      - action: continue
        message: "Flow complete despite cleanup issues"
`,"triggers-example":`# =============================================================================
# Flow Workflow: All Trigger Types Example
# =============================================================================
# This file demonstrates ALL trigger types available in osmedeus workflows.
# Triggers define when/how a workflow should automatically execute.
# Trigger types: cron, event, watch, manual
# =============================================================================

kind: flow
name: triggers-example
description: Demonstrates all trigger types with comprehensive field documentation
tags: triggers, automation, scheduled

# -----------------------------------------------------------------------------
# TRIGGERS SECTION
# Define automatic execution triggers for this workflow
# Multiple triggers can be defined; any triggered condition will start execution
# =============================================================================
trigger:
  # ===========================================================================
  # TRIGGER TYPE: cron
  # Schedule-based execution using cron expressions
  # ===========================================================================
  - # name: Identifier for this trigger (for logging and management)
    name: daily-scan

    # on: Trigger type - cron, event, watch, or manual
    on: cron

    # schedule: Cron expression defining when to run
    # Format: minute hour day-of-month month day-of-week
    # Examples:
    #   "0 0 * * *"     - Every day at midnight
    #   "0 */6 * * *"   - Every 6 hours
    #   "0 9 * * 1-5"   - 9 AM on weekdays
    #   "0 0 1 * *"     - First day of every month at midnight
    schedule: "0 2 * * *"  # Every day at 2 AM

    # input: Defines where the target input comes from for scheduled runs
    input:
      # type: Input source type - file, event_data, function, or param
      type: file

      # path: For "file" type - path to file containing targets (one per line)
      path: "/data/targets/active-targets.txt"

    # enabled: Whether this trigger is active
    # true = trigger is active and will fire
    # false = trigger is defined but disabled
    enabled: true

  # ---------------------------------------------------------------------------
  # Cron trigger with function-based input
  # ---------------------------------------------------------------------------
  - name: weekly-full-scan
    on: cron
    schedule: "0 0 * * 0"  # Every Sunday at midnight

    input:
      # type: function - Generate input dynamically using a function
      type: function

      # function: JavaScript function to generate/retrieve targets
      # Can use built-in functions like db queries, API calls, etc.
      function: 'get_targets_from_db("scope:production")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: event
  # Event-driven execution based on system events
  # Events follow topic format: <component>.<event_type>
  # ===========================================================================
  - name: webhook-trigger
    on: event

    # event: Event configuration for event triggers
    event:
      # topic: Event topic to subscribe to
      # Common topics:
      #   webhook.received    - External webhook received
      #   assets.new          - New asset discovered
      #   assets.changed      - Asset data changed
      #   db.change           - Database record changed
      #   watch.files         - File system change detected
      topic: webhook.received

      # filters: JavaScript expressions to filter events
      # Event data available as 'event' object with fields:
      #   event.name      - Event name
      #   event.source    - Event source
      #   event.data      - JSON payload (string)
      #   event.data_type - Type of data
      # All filters must evaluate to true for trigger to fire
      filters:
        - 'event.source == "github"'
        - 'event.name == "push"'

    # input: How to extract target from event data
    input:
      # type: event_data - Extract from event payload
      type: event_data

      # field: JSON path to extract from event.data
      # Uses dot notation for nested fields
      field: "repository.html_url"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger for new asset discovery
  # ---------------------------------------------------------------------------
  - name: new-asset-scan
    on: event

    event:
      topic: assets.new

      filters:
        # Filter for specific asset types
        - 'event.data_type == "subdomain"'
        # Filter by source tool
        - 'event.source == "subfinder" || event.source == "amass"'

    input:
      type: event_data
      field: "hostname"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger with function-based input extraction
  # ---------------------------------------------------------------------------
  - name: vuln-alert-trigger
    on: event

    event:
      topic: webhook.received

      filters:
        - 'event.name == "vulnerability_alert"'
        - 'JSON.parse(event.data).severity == "critical"'

    input:
      # type: function - Use function to parse/transform event data
      type: function

      # function: Transform event data to target format
      function: 'jq("{{event.data}}", ".affected_host")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: watch
  # File system watch - triggers when files change
  # ===========================================================================
  - name: targets-file-watch
    on: watch

    # path: File or directory path to watch for changes
    # Supports glob patterns in some implementations
    path: "/data/targets/new-targets.txt"

    # input: How to get targets when file changes
    input:
      type: file
      path: "/data/targets/new-targets.txt"

    enabled: true

  # ---------------------------------------------------------------------------
  # Watch trigger on directory
  # ---------------------------------------------------------------------------
  - name: input-directory-watch
    on: watch

    path: "/data/incoming/"

    input:
      # type: function - Process newly added files
      type: function
      function: 'get_new_files("/data/incoming/", "*.txt")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: manual
  # Explicit manual trigger control
  # Used to enable/disable CLI execution for this workflow
  # ===========================================================================
  - name: manual-execution
    on: manual

    # For manual triggers, enabled controls whether CLI can run this workflow
    # enabled: true  - Allow: osmedeus run -f triggers-example -t target
    # enabled: false - Block CLI execution (only scheduled/event triggers work)
    enabled: true

    # input: Default input for manual execution
    # This is optional; CLI -t flag overrides this
    input:
      # type: param - Use a parameter as input
      type: param

      # name: Parameter name to use as target
      name: Target

  # ---------------------------------------------------------------------------
  # Disabled manual trigger example
  # This workflow can ONLY be triggered via cron/events, not CLI
  # ---------------------------------------------------------------------------
  # Uncomment to see the effect:
  # - name: block-manual
  #   on: manual
  #   enabled: false

# -----------------------------------------------------------------------------
# PARAMS SECTION
# -----------------------------------------------------------------------------
params:
  - name: scan_type
    default: "standard"

  - name: threads
    default: "10"

# -----------------------------------------------------------------------------
# MODULES SECTION
# The actual workflow steps to execute when any trigger fires
# -----------------------------------------------------------------------------
modules:
  - name: initial-recon
    path: modules/recon.yaml
    params:
      threads: "{{threads}}"

  - name: scanning
    path: modules/scan.yaml
    depends_on:
      - initial-recon
    params:
      scan_type: "{{scan_type}}"

  - name: reporting
    path: modules/report.yaml
    depends_on:
      - scanning

    on_success:
      - action: notify
        notify: "Triggered scan completed for {{Target}}"
        # condition: Only notify for certain triggers
        condition: 'true'

      - action: export
        name: completed_at
        value: "{{currentDate()}}"
`,"docker-runner-example":`# =============================================================================
# Module Workflow: Docker Runner Configuration Example
# =============================================================================
# This file demonstrates all Docker runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: docker-runner-example
description: Demonstrates Docker runner configuration with all available fields
tags: docker, runner, container

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: docker

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # DOCKER-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # image: Docker image to use (required for docker runner)
  # Format: registry/image:tag or just image:tag
  image: ubuntu:22.04

  # env: Environment variables to set inside the container
  # Map of VAR_NAME: value
  env:
    MY_VAR: my-value
    API_KEY: "{{api_key}}"  # Can use template variables
    THREADS: "{{threads}}"

  # volumes: Volume mounts in docker format
  # Format: host_path:container_path[:options]
  # Options: ro (read-only), rw (read-write)
  volumes:
    - "/tmp/osmedeus:/data"
    - "{{Output}}:/output"
    - "/etc/hosts:/etc/hosts:ro"

  # network: Docker network mode
  # Options: bridge (default), host, none, container:<name>, or network name
  network: host

  # persistent: Container lifecycle mode
  # true = reuse the same container across steps (faster, state preserved)
  # false = ephemeral, create new container per step (isolated, clean state)
  persistent: true

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory inside the container/remote
  # Commands will execute in this directory
  workdir: /app

params:
  - name: api_key
    default: "demo-key"

  - name: threads
    default: "5"

steps:
  # ===========================================================================
  # Step using workflow-level runner (docker with ubuntu:22.04)
  # ===========================================================================
  - name: use-workflow-runner
    type: bash
    log: "Running in workflow-level Docker container"
    command: 'echo "Running inside ubuntu:22.04 container"'

  # ===========================================================================
  # Step with per-step Docker runner override
  # Uses different image than workflow-level config
  # ===========================================================================
  - name: step-with-runner-override
    type: bash
    log: "Running in step-specific Docker container"

    # step_runner: Override runner type for this step only
    # Options: host, docker, ssh
    step_runner: docker

    # step_runner_config: Override runner configuration for this step
    # Same structure as runner_config but applies only to this step
    step_runner_config:
      # Use a different image for this specific step
      image: python:3.11-slim

      env:
        PYTHONPATH: /app

      volumes:
        - "{{Output}}:/output:rw"

      network: bridge

      persistent: false

      workdir: /app

    command: 'python3 -c "print(\\"Running in Python container\\")"'

  # ===========================================================================
  # Remote-bash step type with Docker (explicit remote-bash type)
  # remote-bash is specifically for executing commands in remote environments
  # ===========================================================================
  - name: remote-bash-docker
    # type: remote-bash is specifically for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution in Docker"

    # step_runner: Required for remote-bash type - specifies execution environment
    # Must be "docker" or "ssh"
    step_runner: docker

    step_runner_config:
      image: alpine:latest
      workdir: /tmp

    # command/commands/parallel_commands: Same as bash step
    command: 'echo "Hello from Alpine container" > /tmp/output.txt'

    # step_remote_file: File path on remote (inside container) to copy after execution
    # This file will be copied from the container to the host
    step_remote_file: /tmp/output.txt

    # host_output_file: Local path where the remote file will be copied
    # Template variables are supported
    host_output_file: "{{Output}}/docker-output.txt"

  # ===========================================================================
  # Parallel commands in Docker container
  # ===========================================================================
  - name: docker-parallel-commands
    type: bash
    log: "Running parallel commands in Docker"
    step_runner: docker
    step_runner_config:
      image: ubuntu:22.04
      persistent: true

    parallel_commands:
      - 'sleep 2 && echo "Parallel job A completed"'
      - 'sleep 1 && echo "Parallel job B completed"'
      - 'sleep 3 && echo "Parallel job C completed"'

  # ===========================================================================
  # Foreach loop executing in Docker
  # ===========================================================================
  - name: docker-foreach
    type: foreach
    log: "Processing items in Docker containers"
    input: "{{Output}}/targets.txt"
    variable: target
    threads: 3

    step:
      name: process-in-docker
      type: bash
      step_runner: docker
      step_runner_config:
        image: curlimages/curl:latest
        network: host
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[target]]"'
      exports:
        http_status: "{{stdout}}"

  # ===========================================================================
  # Step running on host (override workflow's docker runner)
  # ===========================================================================
  - name: run-on-host
    type: bash
    log: "Running on host machine (overriding workflow runner)"

    # Override to run locally instead of in container
    step_runner: host

    command: 'echo "This runs directly on the host machine"'

  # ===========================================================================
  # Docker step with all structured arguments
  # ===========================================================================
  - name: docker-with-args
    type: bash
    log: "Docker step with structured arguments"
    step_runner: docker
    step_runner_config:
      image: nuclei:latest
      volumes:
        - "{{Output}}:/output"
        - "/root/nuclei-templates:/templates:ro"
      workdir: /output

    command: nuclei
    speed_args: '-rate-limit 100 -c {{threads}}'
    config_args: '-t /templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /output/nuclei-results.txt'

    step_remote_file: /output/nuclei-results.txt
    host_output_file: "{{Output}}/nuclei-results.txt"

    exports:
      nuclei_output: "{{Output}}/nuclei-results.txt"
`,"ssh-runner-example":`# =============================================================================
# Module Workflow: SSH Runner Configuration Example
# =============================================================================
# This file demonstrates all SSH runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: ssh-runner-example
description: Demonstrates SSH runner configuration with all available fields
tags: ssh, runner, remote

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: ssh

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # SSH-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # host: SSH hostname or IP address (required for ssh runner)
  # Can use template variables for dynamic targeting
  host: "{{ssh_host}}"

  # port: SSH port number
  # Default: 22
  port: 22

  # user: SSH username for authentication
  user: "{{ssh_user}}"

  # key_file: Path to SSH private key file for key-based authentication
  # Preferred over password authentication for security
  key_file: "{{ssh_key_path}}"

  # password: SSH password for password-based authentication
  # WARNING: Not recommended - use key_file instead when possible
  # Can use template variables or environment references
  # password: "{{ssh_password}}"

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory on the remote machine
  # Commands will execute in this directory
  workdir: /home/scanner/workspace

params:
  - name: ssh_host
    default: "192.168.1.100"
    required: true

  - name: ssh_user
    default: "scanner"
    required: true

  - name: ssh_key_path
    default: "~/.ssh/id_rsa"

  - name: threads
    default: "10"

steps:
  # ===========================================================================
  # Step using workflow-level SSH runner
  # ===========================================================================
  - name: setup-remote-workspace
    type: bash
    log: "Setting up workspace on remote SSH server"
    command: 'mkdir -p /home/scanner/workspace/results && echo "Workspace ready"'

  # ===========================================================================
  # Remote-bash step type with SSH (explicit remote-bash type)
  # remote-bash is specifically designed for remote execution scenarios
  # ===========================================================================
  - name: remote-bash-ssh
    # type: remote-bash is explicitly for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution via SSH"

    # step_runner: Required for remote-bash type - must be "docker" or "ssh"
    step_runner: ssh

    # step_runner_config: SSH configuration (inherits from workflow if not set)
    # Omitting this uses workflow-level runner_config
    step_runner_config:
      host: "{{ssh_host}}"
      port: 22
      user: "{{ssh_user}}"
      key_file: "{{ssh_key_path}}"
      workdir: /tmp

    # command: Command to execute on remote server
    command: 'hostname && whoami && pwd > /tmp/remote-info.txt'

    # step_remote_file: File on remote server to copy back to local host
    # This is useful for retrieving results from remote execution
    step_remote_file: /tmp/remote-info.txt

    # host_output_file: Local path where remote file will be copied
    host_output_file: "{{Output}}/remote-info.txt"

    exports:
      remote_file: "{{Output}}/remote-info.txt"

  # ===========================================================================
  # Step overriding SSH connection to different server
  # ===========================================================================
  - name: connect-to-secondary-server
    type: bash
    log: "Connecting to secondary server"

    # Override workflow runner with different SSH target
    step_runner: ssh

    step_runner_config:
      host: "192.168.1.101"  # Different server
      port: 2222             # Non-standard port
      user: admin
      key_file: "~/.ssh/secondary_key"
      workdir: /opt/scanner

    command: 'echo "Connected to secondary server" && uptime'

  # ===========================================================================
  # Multiple sequential commands via SSH
  # ===========================================================================
  - name: ssh-multiple-commands
    type: bash
    log: "Running multiple commands on remote"

    # commands: List of commands executed sequentially on remote
    commands:
      - 'echo "Step 1: Checking system"'
      - 'df -h'
      - 'echo "Step 2: Checking memory"'
      - 'free -m'
      - 'echo "Step 3: Checking processes"'
      - 'ps aux | head -10'

    std_file: "{{Output}}/system-check.txt"

  # ===========================================================================
  # Parallel commands on SSH (run concurrently on remote)
  # ===========================================================================
  - name: ssh-parallel-commands
    type: bash
    log: "Running parallel commands on remote SSH server"

    parallel_commands:
      - 'nmap -sS -p 80 {{Target}} > /tmp/port80.txt'
      - 'nmap -sS -p 443 {{Target}} > /tmp/port443.txt'
      - 'nmap -sS -p 22 {{Target}} > /tmp/port22.txt'

  # ===========================================================================
  # Run tool with structured arguments via SSH
  # ===========================================================================
  - name: ssh-nuclei-scan
    type: bash
    log: "Running nuclei scan via SSH"
    timeout: 3600

    command: nuclei
    speed_args: '-rate-limit 50 -c {{threads}}'
    config_args: '-t ~/nuclei-templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /home/scanner/workspace/nuclei-results.json -json'

    step_remote_file: /home/scanner/workspace/nuclei-results.json
    host_output_file: "{{Output}}/nuclei-results.json"

    exports:
      scan_results: "{{Output}}/nuclei-results.json"

  # ===========================================================================
  # Foreach loop with SSH execution
  # Processes multiple targets on remote server
  # ===========================================================================
  - name: ssh-foreach-targets
    type: foreach
    log: "Processing targets via SSH"

    # input: File containing targets (one per line)
    input: "{{Output}}/targets.txt"

    # variable: Loop variable accessed as [[variable]] in inner step
    variable: current_target

    # threads: Number of concurrent SSH executions
    threads: 5

    step:
      name: probe-target
      type: bash
      # Inner step inherits workflow-level SSH runner
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[current_target]]" 2>/dev/null || echo "failed"'
      exports:
        probe_result: "{{stdout}}"

  # ===========================================================================
  # Step running on local host (override workflow's SSH runner)
  # Useful for local processing of results retrieved from remote
  # ===========================================================================
  - name: process-results-locally
    type: bash
    log: "Processing results on local host"

    # Override to run locally instead of via SSH
    step_runner: host

    command: 'cat "{{Output}}/nuclei-results.json" | jq -r ".info.severity" | sort | uniq -c'

    exports:
      severity_summary: "{{stdout}}"

  # ===========================================================================
  # Function step (always runs locally, regardless of workflow runner)
  # Note: Function steps execute on the host running osmedeus, not remote
  # ===========================================================================
  - name: log-completion
    type: function
    log: "Logging scan completion"
    function: 'log_info("SSH scan completed for {{Target}}")'

  # ===========================================================================
  # Cleanup step on remote server
  # ===========================================================================
  - name: cleanup-remote
    type: bash
    log: "Cleaning up remote workspace"
    command: 'rm -rf /home/scanner/workspace/temp/* 2>/dev/null; echo "Cleanup complete"'

    on_success:
      - action: log
        message: "Remote cleanup completed successfully"

    on_error:
      - action: continue
        message: "Cleanup failed but continuing workflow"
`,"all-step-types-example":`# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: all-step-types-example

# description: Human-readable description of what this workflow does
description: Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  # name: Parameter identifier used in templates as {{param_name}}
  # default: Default value if not provided via CLI
  # required: If true, workflow fails without this value
  # generator: Function to generate value, e.g., uuid(), currentDate(), getEnvVar("KEY")
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"  # Can reference built-in variables
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()  # Generates a unique ID automatically

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  # commands: List of binaries/commands that must exist in PATH
  commands:
    - echo
    - curl

  # files: List of files/directories that must exist
  files:
    - /tmp

  # variables: Define variable requirements with type validation
  # Types: domain, path, number, file, string
  variables:
    - name: Target
      type: string
      required: true

  # functions_conditions: JavaScript expressions that must evaluate to true
  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  # name: Display name for the report
  # path: File path (can use templates like {{Output}})
  # type: Format type - text, csv, json, markdown, etc.
  # description: Human-readable description
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  # skip_workspace: Equivalent to --disable-workspace-creation
  skip_workspace: false

  # disable_notifications: Equivalent to --disable-notification
  disable_notifications: true

  # disable_logging: Equivalent to --disable-logging
  disable_logging: false

  # heuristics_check: Equivalent to --heuristics-check (none, basic, advanced)
  heuristics_check: 'basic'

  # ci_output_format: Equivalent to --ci-output-format
  ci_output_format: false

  # silent: Equivalent to --silent
  silent: false

  # repeat: Equivalent to --repeat
  repeat: false

  # repeat_wait_time: Equivalent to --repeat-wait-time (e.g., 30s, 1h, 2h30m)
  repeat_wait_time: '60s'

  # clean_up_workspace: Equivalent to --clean-up-workspace
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  # ===========================================================================
  # STEP TYPE: bash
  # Execute shell commands on the host (or configured runner)
  # ===========================================================================
  - name: bash-single-command
    # type: Step type - bash, function, parallel-steps, foreach, remote-bash, http, llm
    type: bash

    # pre_condition: JavaScript expression - step only runs if this evaluates to true
    pre_condition: 'true'

    # log: Custom log message displayed when step starts (supports templates)
    log: "Executing single bash command for {{Target}}"

    # timeout: Maximum execution time in seconds (0 = no timeout)
    timeout: 60

    # command: Single command to execute
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'

    # std_file: File path to save stdout/stderr output
    std_file: "{{Output}}/step1-output.txt"

    # exports: Variables to export for subsequent steps
    # Key = variable name, Value = extraction pattern or literal value
    exports:
      step1_result: "completed"

  # ---------------------------------------------------------------------------
  # Bash step with multiple sequential commands
  # ---------------------------------------------------------------------------
  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"

    # commands: List of commands executed sequentially
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  # ---------------------------------------------------------------------------
  # Bash step with parallel commands
  # ---------------------------------------------------------------------------
  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"

    # parallel_commands: List of commands executed concurrently
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  # ---------------------------------------------------------------------------
  # Bash step with structured arguments
  # Arguments are joined in order: command + speed + config + input + output
  # ---------------------------------------------------------------------------
  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"

    command: 'echo'

    # speed_args: Performance-related arguments (e.g., thread count, rate limits)
    speed_args: '-n'

    # config_args: Configuration arguments (e.g., config file paths)
    config_args: ''

    # input_args: Input-related arguments (e.g., input file, target)
    input_args: '"Structured arguments test"'

    # output_args: Output-related arguments (e.g., output file, format)
    output_args: ''

  # ===========================================================================
  # STEP TYPE: function
  # Execute built-in utility functions via Otto JavaScript runtime
  # ===========================================================================
  - name: function-single
    type: function
    log: "Executing single function"

    # function: Single function call (JavaScript expression)
    function: 'log_info("Processing {{Target}} in function step")'

  # ---------------------------------------------------------------------------
  # Function step with multiple sequential functions
  # ---------------------------------------------------------------------------
  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"

    # functions: List of functions executed sequentially
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  # ---------------------------------------------------------------------------
  # Function step with parallel functions
  # ---------------------------------------------------------------------------
  - name: function-parallel
    type: function
    log: "Executing functions in parallel"

    # parallel_functions: List of functions executed concurrently
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  # ===========================================================================
  # STEP TYPE: parallel-steps
  # Execute multiple complete steps in parallel
  # ===========================================================================
  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"

    # parallel_steps: List of Step objects executed concurrently
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'

      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'

      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  # ===========================================================================
  # STEP TYPE: foreach
  # Iterate over input lines, executing inner step for each
  # ===========================================================================
  - name: foreach-example
    type: foreach
    log: "Iterating over items"

    # input: File path or direct content to iterate over (one item per line)
    input: "{{Output}}/items.txt"

    # variable: Name for the loop variable, accessed as [[variable]] in inner step
    variable: item

    # threads: Number of concurrent iterations (default: 1 = sequential)
    threads: 5

    # step: The inner step to execute for each item (single Step object)
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  # ===========================================================================
  # STEP TYPE: http
  # Make HTTP requests to external APIs
  # ===========================================================================
  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30

    # url: Target URL for the request (required for http type)
    url: "https://httpbin.org/post"

    # method: HTTP method - GET, POST, PUT, DELETE, PATCH, etc.
    method: POST

    # headers: Map of HTTP headers to send
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value

    # request_body: Request body content (typically JSON for POST/PUT)
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }

    exports:
      http_response: "{{response.body}}"

  # ===========================================================================
  # STEP TYPE: llm
  # Make LLM API calls for AI-powered processing
  # ===========================================================================
  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120

    # messages: Conversation messages for chat completion
    # role: system, user, assistant, or tool
    # content: Message text (can be string or multimodal array)
    messages:
      - role: system
        content: "You are a security analysis assistant."

      - role: user
        # content can be a simple string or complex multimodal content
        content: "Analyze this target: {{Target}}"

    # tools: Function tools available to the LLM
    tools:
      - type: function  # Currently only "function" type supported
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          # parameters: JSON Schema defining function parameters
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target

    # tool_choice: How the model should choose tools
    # Can be: "auto", "none", "required", or {"type": "function", "function": {"name": "fn_name"}}
    tool_choice: auto

    # llm_config: Step-level LLM configuration overrides
    llm_config:
      # provider: Specific provider to use (overrides rotation)
      provider: openai

      # model: Model override for this step
      model: gpt-4

      # Generation parameters
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0

      # Request settings
      timeout: "60s"
      max_retries: 3
      stream: false

      # response_format: Control output format
      # type: "text", "json_object", or "json_schema"
      response_format:
        type: json_object

    # extra_llm_parameters: Additional provider-specific parameters
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0

    exports:
      llm_analysis: "{{response.content}}"

  # ---------------------------------------------------------------------------
  # LLM step for embeddings
  # ---------------------------------------------------------------------------
  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"

    # is_embedding: Flag to indicate this is an embedding request
    is_embedding: true

    # embedding_input: List of texts to generate embeddings for
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"

    llm_config:
      model: text-embedding-3-small

    exports:
      embeddings: "{{response.embeddings}}"

  # ===========================================================================
  # COMMON STEP FIELDS: on_success, on_error, decision
  # These fields are available on ALL step types
  # ===========================================================================
  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'

    # on_success: Actions to execute when step succeeds
    on_success:
      # action: Handler type - log, abort, continue, export, run, notify
      - action: log
        message: "Step completed successfully for {{Target}}"

      - action: export
        # name: Variable name to export
        name: success_flag
        # value: Value to export (can be string, number, or template)
        value: "true"

      - action: notify
        # notify: Notification message
        notify: "Step succeeded for {{Target}}"

      - action: run
        # type: Step type to run (bash or function)
        type: bash
        command: 'echo "Running follow-up command"'

      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'

    # on_error: Actions to execute when step fails
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        # condition: Only execute this action if condition evaluates to true
        condition: 'true'

      - action: notify
        notify: "Error in workflow for {{Target}}"

      # abort: Stops workflow execution immediately
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'  # Only abort under specific conditions

      # continue: Allows workflow to continue despite error
      - action: continue
        message: "Continuing despite error"

    # decision: Conditional routing to other steps or workflow end
    decision:
      # condition: JavaScript expression to evaluate
      # next: Step name to jump to, or "_end" to finish workflow
      - condition: '{{success_flag}} == "true"'
        next: final-step

      - condition: '{{success_flag}} != "true"'
        next: _end  # Special value to end workflow

  # ---------------------------------------------------------------------------
  # Final step
  # ---------------------------------------------------------------------------
  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`,"mock-all-step-types-example":`# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: mock-all-step-types-example

# description: Human-readable description of what this workflow does
description: Mock Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - echo
    - curl

  files:
    - /tmp

  variables:
    - name: Target
      type: string
      required: true

  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: true
  disable_logging: false
  heuristics_check: 'basic'
  ci_output_format: false
  silent: false
  repeat: false
  repeat_wait_time: '60s'
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  - name: bash-single-command
    type: bash
    pre_condition: 'true'
    log: "Executing single bash command for {{Target}}"
    timeout: 60
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'
    std_file: "{{Output}}/step1-output.txt"
    exports:
      step1_result: "completed"

  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"
    command: 'echo'
    speed_args: '-n'
    config_args: ''
    input_args: '"Structured arguments test"'
    output_args: ''

  - name: function-single
    type: function
    log: "Executing single function"
    function: 'log_info("Processing {{Target}} in function step")'

  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  - name: function-parallel
    type: function
    log: "Executing functions in parallel"
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'
      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'
      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  - name: foreach-example
    type: foreach
    log: "Iterating over items"
    input: "{{Output}}/items.txt"
    variable: item
    threads: 5
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30
    url: "https://httpbin.org/post"
    method: POST
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }
    exports:
      http_response: "{{response.body}}"

  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120
    messages:
      - role: system
        content: "You are a security analysis assistant."
      - role: user
        content: "Analyze this target: {{Target}}"
    tools:
      - type: function
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target
    tool_choice: auto
    llm_config:
      provider: openai
      model: gpt-4
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0
      timeout: "60s"
      max_retries: 3
      stream: false
      response_format:
        type: json_object
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0
    exports:
      llm_analysis: "{{response.content}}"

  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"
    is_embedding: true
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"
    llm_config:
      model: text-embedding-3-small
    exports:
      embeddings: "{{response.embeddings}}"

  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'
    on_success:
      - action: log
        message: "Step completed successfully for {{Target}}"
      - action: export
        name: success_flag
        value: "true"
      - action: notify
        notify: "Step succeeded for {{Target}}"
      - action: run
        type: bash
        command: 'echo "Running follow-up command"'
      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        condition: 'true'
      - action: notify
        notify: "Error in workflow for {{Target}}"
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'
      - action: continue
        message: "Continuing despite error"
    decision:
      - condition: '{{success_flag}} == "true"'
        next: final-step
      - condition: '{{success_flag}} != "true"'
        next: _end

  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`};e.s(["MOCK_WORKFLOW_YAMLS",0,t])},51673,e=>{"use strict";var t=e.i(55161),n=e.i(62280),o=e.i(72536),r=e.i(37364),i=e.i(57763);function a(){try{let e=window.localStorage.getItem("osmedeus_custom_workflows");if(!e)return{};let t=JSON.parse(e);if(!t||"object"!=typeof t)return{};let n={};return Object.entries(t).forEach(([e,t])=>{"string"!=typeof t||t.trim()&&(n[String(e)]=t)}),n}catch{return{}}}function s(){let e=a();return{...r.MOCK_WORKFLOW_YAMLS,...e}}function l(){let e=[];return Object.entries(r.MOCK_WORKFLOW_YAMLS).forEach(([t,n])=>{"string"==typeof n&&n.trim()&&e.push({id:t,content:n,source:"builtin"})}),Object.entries(a()).forEach(([t,n])=>{"string"==typeof n&&n.trim()&&e.push({id:t,content:n,source:"custom"})}),e}function u(e){let t=s()[e];if("string"==typeof t&&t.trim())return t;for(let{id:t,content:n}of l().slice().reverse()){let o={};try{o=i.default.load(n)||{}}catch{o={}}let r="string"==typeof o?.name?o.name.trim():"";if(r&&r===e||t===e)return n}return null}function c(){let e=l(),t=new Map,n=[];return e.forEach(({id:e,content:o,source:r})=>{let i=f(e,o),a=(i.name||"").trim()||e,s=t.get(a);if(!s){t.set(a,{wf:i,source:r}),n.push(a);return}"builtin"===s.source&&"custom"===r&&t.set(a,{wf:i,source:r})}),n.map(e=>t.get(e).wf)}function p(e){return Array.isArray(e)?e.filter(e=>"string"==typeof e).map(e=>e.trim()).filter(Boolean):"string"==typeof e?e.split(",").map(e=>e.trim()).filter(Boolean):[]}function d(e){let t=parseInt((e instanceof Error?e.message:"").split(":")[0]||"0",10);return Number.isFinite(t)?t:0}function m(){(0,o.setDemoMode)(!0)}function f(e,t){let n,o={};try{o=i.default.load(t)||{}}catch{o={}}let r=Array.isArray(o?.steps)?o.steps:[],a=Array.isArray(o?.modules)?o.modules:[],s=o?.kind==="flow"?"flow":"module",l="string"==typeof o?.name?o.name:e,u="string"==typeof o?.description?o.description:"",c=((n=new Set(p(o?.tags))).add("mock-data"),Array.from(n)),d=Array.isArray(o?.params)?o.params:[];return{name:l,kind:s,description:u,tags:c,file_path:"",params:d,required_params:d.filter(e=>e?.required).map(e=>e?.name??""),step_count:r.length,module_count:a.length,checksum:"",indexed_at:new Date().toISOString()}}function h(){let e=new Set;return Object.values(s()).forEach(t=>{try{let n=i.default.load(t)||{};p(n?.tags).forEach(t=>e.add(t))}catch{}}),e.add("mock-data"),Array.from(e.values()).sort()}async function g(){if((0,o.isDemoMode)())return c();let e=await t.http.get(`${n.API_PREFIX}/workflows`);return(e.data?.data||[]).map(e=>({name:e.name??"",kind:"flow"===e.kind?"flow":"module",description:e.description??"",tags:Array.isArray(e.tags)?e.tags:[],file_path:e.file_path??"",params:Array.isArray(e.params)?e.params:[],required_params:Array.isArray(e.required_params)?e.required_params:[],step_count:e.step_count??0,module_count:e.module_count??0,checksum:e.checksum??"",indexed_at:e.indexed_at??""}))}async function y(e={}){let t=c().filter(t=>{if(e.kind&&t.kind!==e.kind)return!1;if(e.tags&&e.tags.length>0){let n=new Set((t.tags||[]).map(e=>String(e)));if(!e.tags.some(e=>n.has(e)))return!1}if(e.search&&e.search.trim()){let n=e.search.trim().toLowerCase();if(!`${t.name??""} ${t.description??""} ${(t.tags||[]).join(" ")}`.toLowerCase().includes(n))return!1}return!0}),n="number"==typeof e.offset?e.offset:0,o="number"==typeof e.limit?e.limit:t.length;return{items:t.slice(Math.max(0,n),Math.max(0,n)+Math.max(0,o)),pagination:{total:t.length,offset:n,limit:o}}}async function w(e={}){if((0,o.isDemoMode)()){let t=(await g()).filter(t=>{if(e.kind&&t.kind!==e.kind)return!1;if(e.tags&&e.tags.length>0){let n=new Set((t.tags||[]).map(e=>String(e)));if(!e.tags.some(e=>n.has(e)))return!1}if(e.search&&e.search.trim()){let n=e.search.trim().toLowerCase();if(!`${t.name??""} ${t.description??""} ${(t.tags||[]).join(" ")}`.toLowerCase().includes(n))return!1}return!0}),n="number"==typeof e.offset?e.offset:0,o="number"==typeof e.limit?e.limit:t.length;return{items:t.slice(Math.max(0,n),Math.max(0,n)+Math.max(0,o)),pagination:{total:t.length,offset:n,limit:o}}}let r={};e.source&&(r.source=e.source),e.tags&&e.tags.length>0&&(r.tags=e.tags.join(",")),e.kind&&(r.kind=e.kind),e.search&&(r.search=e.search),"number"==typeof e.offset&&(r.offset=e.offset),"number"==typeof e.limit&&(r.limit=e.limit);try{let e=await t.http.get(`${n.API_PREFIX}/workflows`,{params:r}),o=e.data?.data||[],i=e.data?.pagination||{total:o.length,offset:0,limit:o.length},a=o.map(e=>({name:e.name??"",kind:"flow"===e.kind?"flow":"module",description:e.description??"",tags:Array.isArray(e.tags)?e.tags.map(e=>String(e)):[],file_path:e.file_path??"",params:Array.isArray(e.params)?e.params:[],required_params:Array.isArray(e.required_params)?e.required_params:[],step_count:e.step_count??0,module_count:e.module_count??0,checksum:e.checksum??"",indexed_at:e.indexed_at??""}));return{items:a,pagination:{total:Number(i.total)||a.length,offset:Number(i.offset)||0,limit:Number(i.limit)||a.length}}}catch(t){if(0===d(t))return m(),y({kind:e.kind,tags:e.tags,search:e.search,offset:e.offset,limit:e.limit});throw t}}async function b(e){if((0,o.isDemoMode)()){let t=u(e);return t?f(e,t):null}try{let o=(await t.http.get(`${n.API_PREFIX}/workflows/${encodeURIComponent(e)}`,{params:{json:!0}})).data;return{name:o.name??"",kind:"flow"===o.kind?"flow":"module",description:o.description??"",tags:Array.isArray(o.tags)?o.tags:[],file_path:o.file_path??"",params:Array.isArray(o.params)?o.params:[],required_params:Array.isArray(o.required_params)?o.required_params:[],step_count:Array.isArray(o.steps)?o.steps.length:o.step_count??0,module_count:o.module_count??0,checksum:o.checksum??"",indexed_at:o.indexed_at??""}}catch(n){let t=d(n);if(404===t)throw Error("WORKFLOW_NOT_FOUND");if(401===t)throw Error("UNAUTHORIZED");if(0===t){m();let t=u(e);return t?f(e,t):null}throw Error("REQUEST_FAILED")}}async function _(e){if((0,o.isDemoMode)())return u(e);try{let o=await t.http.get(`${n.API_PREFIX}/workflows/${encodeURIComponent(e)}`,{responseType:"text"});return"string"==typeof o.data?o.data:o.data?.yaml??null}catch(n){let t=d(n);if(404===t)throw Error("WORKFLOW_NOT_FOUND");if(401===t)throw Error("UNAUTHORIZED");if(0===t)return m(),u(e);throw Error("REQUEST_FAILED")}}async function v(){if((0,o.isDemoMode)())return h();try{let e=await t.http.get(`${n.API_PREFIX}/workflows/tags`),o=e.data?.tags||[];return Array.isArray(o)?o.map(e=>String(e)):[]}catch(e){if(0===d(e))return m(),h();throw e}}async function k(e=!1){let o=await t.http.post(`${n.API_PREFIX}/workflows/refresh`,void 0,{params:e?{force:!0}:{}});return{message:o.data?.message||"",added:Number(o.data?.added||0),updated:Number(o.data?.updated||0),removed:Number(o.data?.removed||0),errors:Array.isArray(o.data?.errors)?o.data.errors:[]}}async function x(e,r){if(!e||!r.trim())return!1;if((0,o.isDemoMode)())try{let t=window.localStorage.getItem("osmedeus_custom_workflows"),n=t?JSON.parse(t):{},o=n&&"object"==typeof n?n:{};return o[e]=r,window.localStorage.setItem("osmedeus_custom_workflows",JSON.stringify(o)),!0}catch{return!1}try{let o=e,a="module";try{let e=i.default.load(r)||{};"string"==typeof e?.name&&e.name.trim()&&(o=e.name.trim()),e?.kind==="flow"&&(a="flow")}catch{}let s=new FormData,l=`${o||e}.yaml`,u=new Blob([r],{type:"text/yaml"});return s.append("file",u,l),await t.http.post(`${n.API_PREFIX}/workflow-upload`,s,{headers:{"Content-Type":"multipart/form-data"},params:{kind:a}}),!0}catch(t){if(0===d(t))return(0,o.setDemoMode)(!0),x(e,r);return!1}}e.s(["fetchMockWorkflowsList",()=>y,"fetchWorkflow",()=>b,"fetchWorkflowTags",()=>v,"fetchWorkflowYaml",()=>_,"fetchWorkflows",()=>g,"fetchWorkflowsList",()=>w,"refreshWorkflowIndex",()=>k,"saveWorkflowYaml",()=>x])}]);