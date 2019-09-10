from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class IPSpace(skeleton.Skeleton):
    """docstring for LinkFinding"""

    def banner(self):
        utils.print_banner("Starting IPSpace")
        utils.make_directory(self.options['WORKSPACE'] + '/ipspace')

    # reading ASN and IP space from amass result
    def get_amass(self, command):
        ips, ip_range, asns = [], [], []
        output_path = command.get('output_path')
        content = utils.just_read(output_path, get_list=True)

        if content:
            ip_range.extend(content)

        amass_output = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/amass-$OUTPUT/amass.json')
        if not utils.not_empty_file(amass_output):
            return

        real_ips = summary.get_ip(self.options)
        if real_ips:
            ips.extend(real_ips)

        # parsed output from amass result
        jsonl = utils.just_read(amass_output, get_list=True)
        for line in jsonl:
            json_data = utils.get_json(line)
            for item in json_data.get('addresses'):
                ips.append(item.get('ip'))
                ip_range.append(item.get('cidr'))
                asns.append(str(item.get('asn')))

        ips_file = utils.replace_argument(
            self.options, '$WORKSPACE/ipspace/summary-$OUTPUT.txt')
        asn_file = utils.replace_argument(
            self.options, '$WORKSPACE/ipspace/asn-$OUTPUT.txt')
        range_file = utils.replace_argument(
            self.options, '$WORKSPACE/ipspace/range-$OUTPUT.txt')

        utils.just_write(ips_file, "\n".join(utils.unique_list(ips)))
        utils.just_write(asn_file, "\n".join(utils.unique_list(asns)))
        utils.just_write(range_file, "\n".join(utils.unique_list(ip_range)))
