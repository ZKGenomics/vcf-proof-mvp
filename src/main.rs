use flate2::read::MultiGzDecoder;
use std::fs::File;
use std::io::BufReader;
use vcf::{U8Vec, VCFError, VCFHeaderFilterAlt, VCFReader, VCFRecord};

fn main() -> Result<(), VCFError> {
    let mut reader = VCFReader::new(BufReader::new(MultiGzDecoder::new(File::open(
        "./data/ALL.chr22.shapeit2_integrated_snvindels_v2a_27022019.GRCh38.phased.vcf.gz",
    )?)))?;

    // access FILTER contents
    assert_eq!(
        Some(VCFHeaderFilterAlt {
            id: b"PASS",
            description: b"All filters passed"
        }),
        reader.header().filter(b"PASS")
    );

    // prepare VCFRecord object
    let mut vcf_record = reader.empty_record();

    // read one record
    reader.next_record(&mut vcf_record)?;
    println!("Chromosome: {:?}", vcf_record.chromosome);
    println!("Position: {:?}", vcf_record.position);
    println!("Alternative: {:?}", vcf_record.alternative);
    println!("Format: {:?}", vcf_record.format);
    //println!("Genotype: {:?}", vcf_record.genotype);

    Ok(())
}
