// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace WTTech.Aem.Compose.Outputs
{

    [OutputType]
    public sealed class Compose
    {
        /// <summary>
        /// Contents of the AEM Compose YML configuration file.
        /// </summary>
        public readonly string? Config;
        /// <summary>
        /// Script(s) for configuring a launched instance. Must be idempotent as it is executed always when changed. Typically used for installing AEM service packs, setting up replication agents, etc.
        /// </summary>
        public readonly Outputs.InstanceScript? Configure;
        /// <summary>
        /// Script(s) for creating an instance or restoring it from a backup. Typically customized to provide AEM library files (quickstart.jar, license.properties, service packs) from alternative sources (e.g., AWS S3, Azure Blob Storage). Instance recreation is forced if changed.
        /// </summary>
        public readonly Outputs.InstanceScript? Create;
        /// <summary>
        /// Script(s) for deleting a stopped instance.
        /// </summary>
        public readonly Outputs.InstanceScript? Delete;
        /// <summary>
        /// Toggle automatic AEM Compose CLI wrapper download. If set to false, assume the wrapper is present in the data directory.
        /// </summary>
        public readonly bool? Download;
        /// <summary>
        /// Version of AEM Compose tool to use on remote machine.
        /// </summary>
        public readonly string? Version;

        [OutputConstructor]
        private Compose(
            string? config,

            Outputs.InstanceScript? configure,

            Outputs.InstanceScript? create,

            Outputs.InstanceScript? delete,

            bool? download,

            string? version)
        {
            Config = config;
            Configure = configure;
            Create = create;
            Delete = delete;
            Download = download;
            Version = version;
        }
    }
}